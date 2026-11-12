// Reconcile XP achievements in bounded batches. DATABASE_URL selects the database.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"ArthaFreestyle/Arsiva/internal/repository"
)

func main() {
	notify := flag.Bool("notify", false, "create inbox notifications for newly awarded achievements")
	batch := flag.Int("batch-size", 100, "members per batch")
	flag.Parse()
	if *batch < 1 || *batch > 10000 {
		log.Fatal("batch-size must be 1..10000")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	last, awarded, skipped, failed := 0, 0, 0, 0
	for {
		rows, err := db.Query(ctx, `SELECT member_id FROM members WHERE member_id>$1 ORDER BY member_id LIMIT $2`, last, *batch)
		if err != nil {
			log.Fatal(err)
		}
		ids := []int{}
		for rows.Next() {
			var id int
			if err = rows.Scan(&id); err != nil {
				log.Fatal(err)
			}
			ids = append(ids, id)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			log.Fatal(err)
		}
		if len(ids) == 0 {
			break
		}
		for _, id := range ids {
			n, err := repository.ReconcileMember(ctx, db, id, *notify)
			if err != nil {
				failed++
				log.Printf("member=%d error=%v", id, err)
			} else {
				awarded += n
				if n == 0 {
					skipped++
				}
			}
			last = id
		}
	}
	log.Printf("awarded=%d skipped=%d errors=%d", awarded, skipped, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
