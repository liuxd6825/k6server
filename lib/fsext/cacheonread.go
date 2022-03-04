/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2019 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package fsext

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/spf13/afero"
)

// ErrPathNeverRequestedBefore represent an error when path never opened/requested before
var ErrPathNeverRequestedBefore = errors.New("path never requested before")

// CacheOnReadFs is wrapper around afero.CacheOnReadFs with the ability to return the filesystem
// that is used as cache
type CacheOnReadFs struct {
	afero.Fs
	cache      afero.Fs
	lock       *sync.Mutex
	cached     map[string]bool
	cachedOnly bool
}

// OnlyCachedEnabler enables the mode of FS that allows to open
// already opened files (e.g. serve from cache only)
type OnlyCachedEnabler interface {
	AllowOnlyCached()
}

// CacheLayerGetter provide a direct access to a cache layer
type CacheLayerGetter interface {
	GetCachingFs() afero.Fs
}

// NewCacheOnReadFs returns a new CacheOnReadFs
func NewCacheOnReadFs(base, layer afero.Fs, cacheTime time.Duration) afero.Fs {
	return &CacheOnReadFs{
		Fs:    afero.NewCacheOnReadFs(base, layer, cacheTime),
		cache: layer,

		lock:       &sync.Mutex{},
		cachedOnly: false,
		cached:     make(map[string]bool),
	}
}

// GetCachingFs returns the afero.Fs being used for cache
func (c *CacheOnReadFs) GetCachingFs() afero.Fs {
	return c.cache
}

// AllowOnlyCached enables the cached only mode of the CacheOnReadFs
func (c *CacheOnReadFs) AllowOnlyCached() {
	c.lock.Lock()
	c.cachedOnly = true
	c.lock.Unlock()
}

// Open opens file and track the history of opened files
// if CacheOnReadFs is in the opened only mode it should return
// an error if file wasn't open before
func (c *CacheOnReadFs) Open(name string) (afero.File, error) {
	if err := c.checkOrRemember(name); err != nil {
		return nil, err
	}

	return c.Fs.Open(name)
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
// if CacheOnReadFs is in the opened only mode it should return
// an error if path wasn't open before
func (c *CacheOnReadFs) Stat(path string) (os.FileInfo, error) {
	if err := c.checkOrRemember(path); err != nil {
		return nil, err
	}

	return c.Fs.Stat(path)
}

func (c *CacheOnReadFs) checkOrRemember(path string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.cachedOnly {
		c.cached[path] = true
	} else if !c.cached[path] {
		return ErrPathNeverRequestedBefore
	}

	return nil
}
