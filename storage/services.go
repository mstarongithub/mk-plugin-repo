package storage

import "time"

const (
	_SERVICE_KEY_OLD_TOKENS  = "Old token cleaner"
	_SERVICE_KEY_GDPR_EXPIRE = "GDPR 3 month expire cleaner"
)

const _ONE_MONTH_BEFORE = time.Hour * 24 * 30 * -1

// Service worker for cleaning up old and expired tokens
// Runs every hour to delete authentication tokens that have either expired or have been soft deleted
func (storage *Storage) serviceCleanOldTokens(exitChan chan any) {
	storage.serviceWorkersActive.RLock()
	if storage.serviceWorkersActive.Map[_SERVICE_KEY_OLD_TOKENS] {
		return
	}
	storage.serviceWorkersActive.RUnlock()
	storage.serviceWorkersActive.Lock()
	storage.serviceWorkersActive.Map[_SERVICE_KEY_OLD_TOKENS] = true
	storage.serviceWorkersActive.Unlock()
	defer func() {
		storage.serviceWorkersActive.Lock()
		delete(storage.serviceWorkersActive.Map, _SERVICE_KEY_OLD_TOKENS)
		storage.serviceWorkersActive.Unlock()
	}()

	ticker := time.NewTicker(time.Hour)
	for {
		select {
		case stamp := <-ticker.C:
			// Use unscoped here to ensure we don't do soft deletions
			// Old, expired tokens must be gone for good
			storage.db.Unscoped().Where("expires_at < ?", stamp).Delete(&Token{})
			// Now the same for manually deleted tokens
			storage.db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&Token{})
		case <-exitChan:
			ticker.Stop()
			return
		}
	}
}

// Sevice worker for cleaning up soft deleted entries that are older than one month
// Runs every 24h
func (storage *Storage) serviceGdprCleanOldDeletedData(exitChan chan any) {
	storage.serviceWorkersActive.RLock()
	if storage.serviceWorkersActive.Map[_SERVICE_KEY_GDPR_EXPIRE] {
		return
	}
	storage.serviceWorkersActive.RUnlock()
	storage.serviceWorkersActive.Lock()
	storage.serviceWorkersActive.Map[_SERVICE_KEY_GDPR_EXPIRE] = true
	storage.serviceWorkersActive.Unlock()
	defer func() {
		storage.serviceWorkersActive.Lock()
		delete(storage.serviceWorkersActive.Map, _SERVICE_KEY_GDPR_EXPIRE)
		storage.serviceWorkersActive.Unlock()
	}()

	ticker := time.NewTicker(time.Hour * 24)
	for {
		select {
		case stamp := <-ticker.C:
			storage.db.Unscoped().
				Where("deleted_at IS NOT NULL").
				Where("deleted_at < ?", stamp.Add(_ONE_MONTH_BEFORE)).
				Delete(&Account{})
			storage.db.Unscoped().
				Where("deleted_at IS NOT NULL").
				Where("deleted_at < ?", stamp.Add(_ONE_MONTH_BEFORE)).
				Delete(&Plugin{})
			storage.db.Unscoped().
				Where("deleted_at IS NOT NULL").
				Where("deleted_at < ?", stamp.Add(_ONE_MONTH_BEFORE)).
				Delete(&PluginVersion{})
		case <-exitChan:
			ticker.Stop()
			return
		}
	}
}
