import yaml
import sys

def main():
    with open('docs/openapi.yaml', 'r') as f:
        data = yaml.safe_load(f)

    # Add Tag
    if 'tags' not in data:
        data['tags'] = []
    
    # ensure no duplicate Groups tag
    tag_names = [t['name'] for t in data['tags']]
    if 'Groups' not in tag_names:
        data['tags'].append({
            'name': 'Groups',
            'description': 'CRUD manajemen grup dan anggota (🔒 POST/PUT/DELETE/GET = guru, POST join = member)'
        })

    # Add Paths
    if 'paths' not in data:
        data['paths'] = {}

    data['paths']['/api/v1/groups'] = {
        'get': {
            'tags': ['Groups'],
            'summary': 'Get all groups',
            'security': [{'BearerAuth': []}],
            'responses': {'200': {'description': 'Success'}},
            'parameters': [
                {'name': 'page', 'in': 'query', 'schema': {'type': 'integer'}},
                {'name': 'size', 'in': 'query', 'schema': {'type': 'integer'}},
                {'name': 'search', 'in': 'query', 'schema': {'type': 'string'}}
            ]
        },
        'post': {
            'tags': ['Groups'],
            'summary': 'Create group',
            'security': [{'BearerAuth': []}],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    data['paths']['/api/v1/groups/{id}'] = {
        'get': {
            'tags': ['Groups'],
            'summary': 'Get group detail',
            'security': [{'BearerAuth': []}],
            'parameters': [{'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}}],
            'responses': {'200': {'description': 'Success'}}
        },
        'put': {
            'tags': ['Groups'],
            'summary': 'Update group',
            'security': [{'BearerAuth': []}],
            'parameters': [{'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}}],
            'responses': {'200': {'description': 'Success'}}
        },
        'delete': {
            'tags': ['Groups'],
            'summary': 'Delete group',
            'security': [{'BearerAuth': []}],
            'parameters': [{'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}}],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    data['paths']['/api/v1/groups/{id}/invite'] = {
        'post': {
            'tags': ['Groups'],
            'summary': 'Invite members by email',
            'security': [{'BearerAuth': []}],
            'parameters': [{'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}}],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    data['paths']['/api/v1/groups/{id}/invite-link'] = {
        'get': {
            'tags': ['Groups'],
            'summary': 'Generate invite link',
            'security': [{'BearerAuth': []}],
            'parameters': [{'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}}],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    data['paths']['/api/v1/groups/join'] = {
        'post': {
            'tags': ['Groups'],
            'summary': 'Join a group',
            'security': [{'BearerAuth': []}],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    data['paths']['/api/v1/groups/{id}/members'] = {
        'get': {
            'tags': ['Groups'],
            'summary': 'Get group members',
            'security': [{'BearerAuth': []}],
            'parameters': [{'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}}],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    data['paths']['/api/v1/groups/{id}/members/{member_id}'] = {
        'delete': {
            'tags': ['Groups'],
            'summary': 'Remove member',
            'security': [{'BearerAuth': []}],
            'parameters': [
                {'name': 'id', 'in': 'path', 'required': True, 'schema': {'type': 'string'}},
                {'name': 'member_id', 'in': 'path', 'required': True, 'schema': {'type': 'integer'}}
            ],
            'responses': {'200': {'description': 'Success'}}
        }
    }

    # Dump yaml preserving order
    with open('docs/openapi.yaml', 'w') as f:
        yaml.dump(data, f, sort_keys=False, indent=2)

if __name__ == "__main__":
    main()
