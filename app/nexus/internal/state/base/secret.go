//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// UpsertSecret stores or updates a secret at the specified path with the
// provided values. It provides thread-safe access to the underlying key-value
// kv.
//
// Parameters:
//   - path: The location where the secret should be stored
//   - values: A map containing the secret key-value pairs to be stored
func UpsertSecret(path string, values map[string]string) {
	secretStoreMu.Lock()
	secretStore.Put(path, values)
	secretStoreMu.Unlock()

	persist.StoreSecret(secretStore, path)
}

// DeleteSecret deletes one or more versions of a secret at the specified path.
// It acquires a mutex lock before performing the deletion to ensure thread
// safety.
//
// Parameters:
//   - path: The path to the secret to be deleted
//   - versions: A slice of version numbers to delete. If empty, deletes the
//     current version only. Version number 0 is the current version.
func DeleteSecret(path string, versions []int) error {
	secretStoreMu.Lock()
	err := secretStore.Delete(path, versions)
	secretStoreMu.Unlock()

	if err != nil {
		return err
	}

	persist.StoreSecret(secretStore, path)
	return nil
}

// UndeleteSecret restores previously deleted versions of a secret at the
// specified path. It takes a path string identifying the secret's location and
// a slice of version numbers to restore. The function acquires a lock on the
// key-value kv to ensure thread-safe operations during the `undelete` process.
//
// The function operates synchronously and will block until the undelete
// operation is complete. If any specified version numbers don't exist or were
// not previously deleted, those versions will be silently skipped.
//
// Parameters:
//   - path: The path to the secret to be restored
//   - versions: A slice of integer version numbers to restore
//
// Example:
//
//	// Restore versions 1 and 3 of a secret
//	UndeleteSecret("/app/secrets/api-key", []int{1, 3})
func UndeleteSecret(path string, versions []int) error {
	secretStoreMu.Lock()
	err := secretStore.Undelete(path, versions)
	secretStoreMu.Unlock()

	if err != nil {
		return err
	}

	persist.StoreSecret(secretStore, path)
	return nil
}

// GetSecret retrieves a secret from the specified path and version.
// It provides thread-safe read access to the secret kv.
//
// Parameters:
//   - path: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - map[string]string: The secret key-value pairs
//   - bool: Whether the secret was found
func GetSecret(path string, version int) (map[string]string, error) {
	secretStoreMu.RLock()
	secret, err := secretStore.Get(path, version)
	secretStoreMu.RUnlock()

	if err == nil {
		return secret, nil
	}

	cachedSecret := persist.ReadSecret(path, version)
	if cachedSecret == nil {
		return nil, err
	}

	if version == 0 {
		version = cachedSecret.Metadata.CurrentVersion
	}

	secretStoreMu.Lock()
	secretStore.Put(path, cachedSecret.Versions[version].Data)
	secretStoreMu.Unlock()

	return cachedSecret.Versions[version].Data, nil
}

// ImportSecrets imports a set of secrets into the application's memory state.
// Locks the secret store mutex during the operation to ensure thread safety.
func ImportSecrets(secrets map[string]*kv.Value) {
	secretStoreMu.Lock()
	defer secretStoreMu.Unlock()
	secretStore.ImportSecrets(secrets)
}

// GetRawSecret retrieves a secret with metadata from the specified path and
// version. It provides thread-safe read access to the secret kv.
//
// Parameters:
//   - path: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - *kv.Secret: The secret type
//   - bool: Whether the secret was found
func GetRawSecret(path string, version int) (*kv.Value, error) {
	secretStoreMu.RLock()
	secret, err := secretStore.GetRawSecret(path)
	secretStoreMu.RUnlock()

	if err == nil {
		return secret, nil
	}

	cachedSecret := persist.ReadSecret(path, version)
	if cachedSecret == nil {
		return nil, err
	}

	if version == 0 {
		version = cachedSecret.Metadata.CurrentVersion
	}

	secretStoreMu.Lock()
	secretStore.Put(path, cachedSecret.Versions[version].Data)
	secretStoreMu.Unlock()

	return cachedSecret, nil
}
