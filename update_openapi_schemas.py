import yaml

def main():
    with open('docs/openapi.yaml', 'r') as f:
        data = yaml.safe_load(f)

    schemas = data.get('components', {}).get('schemas', {})

    new_schemas = {
        'GroupCreateRequest': {
            'type': 'object',
            'required': ['group_name'],
            'properties': {
                'group_name': {'type': 'string', 'maxLength': 191, 'example': 'Kelas 9A - Matematika'}
            }
        },
        'GroupUpdateRequest': {
            'type': 'object',
            'required': ['group_name'],
            'properties': {
                'group_name': {'type': 'string', 'maxLength': 191, 'example': 'Kelas 9A - Matematika'},
                'thumbnail_asset_id': {'type': 'integer', 'example': 10}
            }
        },
        'GroupInviteEmailRequest': {
            'type': 'object',
            'required': ['emails'],
            'properties': {
                'emails': {
                    'type': 'array',
                    'items': {'type': 'string', 'format': 'email'}
                }
            }
        },
        'GroupJoinRequest': {
            'type': 'object',
            'required': ['invite_token'],
            'properties': {
                'invite_token': {'type': 'string', 'example': 'eyJhbG...'}
            }
        },
        'GroupGuruResponse': {
            'type': 'object',
            'properties': {
                'guru_id': {'type': 'integer', 'example': 1},
                'nip': {'type': 'string', 'example': '198501012010011001'},
                'bidang_ajar': {'type': 'string', 'example': 'Matematika'},
                'username': {'type': 'string', 'example': 'pak_budi'}
            }
        },
        'GroupResponse': {
            'type': 'object',
            'properties': {
                'group_id': {'type': 'string', 'format': 'uuid'},
                'group_name': {'type': 'string', 'example': 'Kelas 9A - Matematika'},
                'thumbnail': {'type': 'string', 'example': 'https://arsiva.app/uploads/thumbnail.webp'},
                'created_by': {'$ref': '#/components/schemas/GroupGuruResponse'},
                'created_at': {'type': 'string', 'example': '2026-04-24 13:00:00'},
                'updated_at': {'type': 'string', 'example': '2026-04-24 13:00:00'},
                'member_count': {'type': 'integer', 'example': 0}
            }
        },
        'GroupMemberResponse': {
            'type': 'object',
            'properties': {
                'member_id': {'type': 'integer', 'example': 1},
                'username': {'type': 'string', 'example': 'andi_siswa'},
                'email': {'type': 'string', 'format': 'email', 'example': 'andi@example.com'},
                'nis': {'type': 'string', 'example': '2024001'},
                'foto_profil': {'type': 'string', 'example': ''},
                'tanggal_bergabung': {'type': 'string', 'example': '2026-04-24 14:00:00'}
            }
        },
        'GroupDetailResponse': {
            'type': 'object',
            'properties': {
                'group_id': {'type': 'string', 'format': 'uuid'},
                'group_name': {'type': 'string', 'example': 'Kelas 9A - Matematika'},
                'thumbnail': {'type': 'string', 'example': 'https://arsiva.app/uploads/thumbnail.webp'},
                'created_by': {'$ref': '#/components/schemas/GroupGuruResponse'},
                'created_at': {'type': 'string', 'example': '2026-04-24 13:00:00'},
                'updated_at': {'type': 'string', 'example': '2026-04-24 13:00:00'},
                'member_count': {'type': 'integer', 'example': 2},
                'members': {
                    'type': 'array',
                    'items': {'$ref': '#/components/schemas/GroupMemberResponse'}
                }
            }
        },
        'GroupInviteResponse': {
            'type': 'object',
            'properties': {
                'invite_token': {'type': 'string'},
                'invite_link': {'type': 'string'},
                'qr_code_data': {'type': 'string'},
                'expires_at': {'type': 'string'}
            }
        },
        'GroupWrapper': {
            'type': 'object',
            'properties': {
                'data': {'$ref': '#/components/schemas/GroupResponse'}
            }
        },
        'GroupDetailWrapper': {
            'type': 'object',
            'properties': {
                'data': {'$ref': '#/components/schemas/GroupDetailResponse'}
            }
        },
        'GroupListWrapper': {
            'type': 'object',
            'properties': {
                'data': {
                    'type': 'array',
                    'items': {'$ref': '#/components/schemas/GroupResponse'}
                },
                'paging': {'$ref': '#/components/schemas/PageMetaData'}
            }
        },
        'GroupInviteWrapper': {
            'type': 'object',
            'properties': {
                'data': {'$ref': '#/components/schemas/GroupInviteResponse'}
            }
        },
        'GroupMemberListWrapper': {
            'type': 'object',
            'properties': {
                'data': {
                    'type': 'array',
                    'items': {'$ref': '#/components/schemas/GroupMemberResponse'}
                }
            }
        },
        'MessageWrapper': {
            'type': 'object',
            'properties': {
                'data': {'type': 'string', 'example': 'Success message'}
            }
        }
    }

    schemas.update(new_schemas)
    if 'components' not in data:
        data['components'] = {}
    data['components']['schemas'] = schemas

    # Update paths to use schemas
    paths = data.get('paths', {})

    if '/api/v1/groups' in paths:
        paths['/api/v1/groups']['get']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupListWrapper'}}}
        }
        paths['/api/v1/groups']['post']['requestBody'] = {
            'required': True,
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupCreateRequest'}}}
        }
        paths['/api/v1/groups']['post']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupWrapper'}}}
        }
        paths['/api/v1/groups']['post']['responses']['201'] = {
            'description': 'Created',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupWrapper'}}}
        }

    if '/api/v1/groups/{id}' in paths:
        paths['/api/v1/groups/{id}']['get']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupDetailWrapper'}}}
        }
        paths['/api/v1/groups/{id}']['put']['requestBody'] = {
            'required': True,
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupUpdateRequest'}}}
        }
        paths['/api/v1/groups/{id}']['put']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupWrapper'}}}
        }
        paths['/api/v1/groups/{id}']['delete']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/MessageWrapper'}}}
        }

    if '/api/v1/groups/{id}/invite' in paths:
        paths['/api/v1/groups/{id}/invite']['post']['requestBody'] = {
            'required': True,
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupInviteEmailRequest'}}}
        }
        paths['/api/v1/groups/{id}/invite']['post']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/MessageWrapper'}}}
        }

    if '/api/v1/groups/{id}/invite-link' in paths:
        paths['/api/v1/groups/{id}/invite-link']['get']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupInviteWrapper'}}}
        }

    if '/api/v1/groups/join' in paths:
        paths['/api/v1/groups/join']['post']['requestBody'] = {
            'required': True,
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupJoinRequest'}}}
        }
        paths['/api/v1/groups/join']['post']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/MessageWrapper'}}}
        }

    if '/api/v1/groups/{id}/members' in paths:
        paths['/api/v1/groups/{id}/members']['get']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/GroupMemberListWrapper'}}}
        }

    if '/api/v1/groups/{id}/members/{member_id}' in paths:
        paths['/api/v1/groups/{id}/members/{member_id}']['delete']['responses']['200'] = {
            'description': 'Success',
            'content': {'application/json': {'schema': {'$ref': '#/components/schemas/MessageWrapper'}}}
        }

    with open('docs/openapi.yaml', 'w') as f:
        yaml.dump(data, f, sort_keys=False, indent=2)

if __name__ == "__main__":
    main()
