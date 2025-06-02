package db

import "sync"

type storage struct {
	sync.Mutex
	permissions map[string]map[string][]string
}

type Repository struct {
	storage *storage
}

func NewRepository(permissions map[string]map[string][]string) *Repository {
	return &Repository{
		storage: &storage{
			permissions: permissions,
		},
	}
}

func (r *Repository) UpdatePermissions(client, scope string, roles []string) error {
	r.storage.Lock()
	defer r.storage.Unlock()

	r.storage.permissions[client][scope] = roles
	return nil
}

func (r *Repository) GetPermissions(client, scope string) []string {
	r.storage.Lock()
	defer r.storage.Unlock()

	clientPerms, ok := r.storage.permissions[client]
	if !ok {
		return []string{}
	}

	roles, ok := clientPerms[scope]
	if !ok {
		return []string{}
	}

	return roles
}
