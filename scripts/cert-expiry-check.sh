#!/usr/bin/env bash
#
# Warn before the arsiva.id TLS certificate expires, and catch the specific
# failure from issue #44: certbot renews the file on disk but nginx keeps
# serving the old certificate out of memory.
#
# Deliberately checks the certificate **on the wire**, not the one on disk --
# the wire is what users actually get, so this catches a failed renewal AND a
# failed reload. It additionally compares the two and reports the drift, which
# surfaces a missed reload the day after a renewal instead of ~40 days later
# when the expiry threshold finally trips.
#
# Install on the VPS (one-time, as root):
#
#   ln -s /home/artha/actions-runner/_work/Arsiva/Arsiva/scripts/cert-expiry-check.sh \
#         /etc/cron.daily/cert-expiry-check
#
# NOTE the missing `.sh` on the link name: run-parts silently skips any file in
# /etc/cron.daily whose name contains a dot, so a link called
# `cert-expiry-check.sh` would never run.
#
# Configure via env (defaults in brackets). Set CERT_ALERT_TO to a real inbox --
# the default goes nowhere useful.
#
#   CERT_DOMAIN      [arsiva.id]        host to probe
#   CERT_PORT        [443]
#   CERT_WARN_DAYS   [20]               alert when fewer days than this remain
#   CERT_FILE        [/etc/letsencrypt/live/$CERT_DOMAIN/fullchain.pem]
#   CERT_ALERT_TO    [ops@arsiva.id]    recipient
#   CERT_ALERT_FROM  [no-reply@arsiva.id]
#
# Exit: 0 healthy, 1 alert raised, 2 could not determine (also alerts).

set -uo pipefail

DOMAIN="${CERT_DOMAIN:-arsiva.id}"
PORT="${CERT_PORT:-443}"
WARN_DAYS="${CERT_WARN_DAYS:-20}"
CERT_FILE="${CERT_FILE:-/etc/letsencrypt/live/${DOMAIN}/fullchain.pem}"
ALERT_TO="${CERT_ALERT_TO:-ops@arsiva.id}"
ALERT_FROM="${CERT_ALERT_FROM:-no-reply@arsiva.id}"

# Mail goes out through the Postfix relay already running on the VPS for the
# app's own mail. sendmail(1) is used rather than mail(1) because bsd-mailx is
# not necessarily installed, whereas Postfix always ships /usr/sbin/sendmail.
notify() {
	subject="$1"
	body="$2"

	# Always leave a trace even if mail delivery fails: cron captures stdout,
	# and syslog survives independently of it.
	printf '%s\n\n%s\n' "$subject" "$body"
	command -v logger >/dev/null 2>&1 && logger -t cert-expiry-check -- "$subject"

	if [ -x /usr/sbin/sendmail ]; then
		# Date and Message-ID are emitted by hand. Gmail rejects mail without a
		# Message-ID at end-of-DATA (550-5.7.1 ... RfcMessageNonCompliant) and
		# Postfix only backfills missing headers when always_add_missing_headers
		# = yes, which is off by default -- see the mailer notes in CLAUDE.md.
		# An alert that is silently bounced is worse than no alert at all.
		/usr/sbin/sendmail -t <<-MAIL
			From: Arsiva Ops <${ALERT_FROM}>
			To: ${ALERT_TO}
			Subject: ${subject}
			Date: $(date -R)
			Message-ID: <cert-expiry-$(date +%s).$$@${DOMAIN}>
			Content-Type: text/plain; charset=UTF-8

			${body}

			--
			${0} on $(hostname -f 2>/dev/null || hostname)
		MAIL
	else
		printf 'cert-expiry-check: /usr/sbin/sendmail not found, no mail sent\n' >&2
	fi
}

# Seconds-since-epoch for an OpenSSL "notAfter" date, or empty on failure.
to_epoch() {
	date -d "$1" +%s 2>/dev/null
}

now=$(date +%s)

# --- what users actually receive -------------------------------------------
# `timeout` matters: a wedged TLS handshake would otherwise hang the cron job
# forever, and openssl has no connect timeout of its own.
wire_end=$(timeout 15 openssl s_client -connect "${DOMAIN}:${PORT}" -servername "${DOMAIN}" \
	</dev/null 2>/dev/null | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2)

if [ -z "${wire_end}" ]; then
	notify "[arsiva] CERT CHECK FAILED - cannot read certificate from ${DOMAIN}:${PORT}" \
		"Could not complete a TLS handshake with ${DOMAIN}:${PORT} within 15s.

That means either the host is down, nginx is not running, or TLS is
misconfigured. Check:

  docker compose ps nginx
  docker compose logs --tail=50 nginx"
	exit 2
fi

wire_epoch=$(to_epoch "${wire_end}")
if [ -z "${wire_epoch}" ]; then
	notify "[arsiva] CERT CHECK FAILED - unparseable expiry date on ${DOMAIN}" \
		"openssl reported notAfter='${wire_end}', which date(1) could not parse."
	exit 2
fi

wire_days=$(( (wire_epoch - now) / 86400 ))

# --- has a renewal landed on disk that nginx has not picked up yet? ---------
# This is the issue #44 signature, and it is worth alerting on immediately
# rather than waiting for the expiry threshold: the fix (a reload) is instant,
# and the certificate on disk is already valid.
if [ -r "${CERT_FILE}" ]; then
	disk_end=$(openssl x509 -noout -enddate -in "${CERT_FILE}" 2>/dev/null | cut -d= -f2)
	disk_epoch=$(to_epoch "${disk_end:-}")

	if [ -n "${disk_epoch}" ] && [ "${disk_epoch}" -gt "${wire_epoch}" ]; then
		notify "[arsiva] STALE CERT - ${DOMAIN} is serving an older certificate than the one on disk" \
			"A renewed certificate is sitting on disk but nginx is still serving the
previous one from memory. nginx parses ssl_certificate once at startup and
never re-reads the file (issue #44).

  served from ${DOMAIN}:${PORT} : expires ${wire_end} (${wire_days} days left)
  on disk (${CERT_FILE}) : expires ${disk_end}

Fix now, gracefully, without dropping connections:

  docker compose exec nginx nginx -s reload

If this recurs, the 6h reload loop in the nginx service of
docker-compose.yml is not running -- confirm with:

  docker inspect -f '{{.Config.Cmd}}' nginx_proxy"
		exit 1
	fi
fi

# --- ordinary expiry threshold ---------------------------------------------
if [ "${wire_days}" -lt 0 ]; then
	notify "[arsiva] CERT EXPIRED - ${DOMAIN} certificate expired $(( -wire_days )) day(s) ago" \
		"${DOMAIN} is serving an EXPIRED certificate. Every visitor is seeing a
browser security warning right now.

  expired: ${wire_end}

Check whether a valid certificate is already on disk, in which case a reload
is all that is needed:

  sudo openssl x509 -noout -dates -in ${CERT_FILE}
  docker compose exec nginx nginx -s reload"
	exit 1
fi

if [ "${wire_days}" -lt "${WARN_DAYS}" ]; then
	notify "[arsiva] CERT EXPIRING - ${DOMAIN} certificate expires in ${wire_days} day(s)" \
		"${DOMAIN} is serving a certificate that expires on ${wire_end}.

Let's Encrypt renews at 30 days remaining, so at ${wire_days} days the renewal
has already been attempted and has not reached users. Check both halves:

  docker compose logs --tail=50 certbot
  docker inspect -f '{{.Config.Cmd}}' nginx_proxy"
	exit 1
fi

printf 'cert-expiry-check: %s OK, expires %s (%d days left)\n' "${DOMAIN}" "${wire_end}" "${wire_days}"
exit 0
