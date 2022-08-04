package automod_legacy

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCachedInvites(t *testing.T) {
	invitesCache.set("foo", 123)
	i, ok := invitesCache.get("foo")
	assert.True(t, ok)
	assert.EqualValues(t, 123, i.guildID)

	time.Sleep(50 * time.Millisecond)
	invitesCache.tick(time.Millisecond)
	_, ok = invitesCache.get("foo")
	assert.False(t, ok)
}

type oldCachedInvite struct {
	CreatedAt time.Time
	Invite    string // we only store the invite code, no need to waste memory on the entire invite data
}

var oldInvitesCache = struct {
	sync.RWMutex

	CacheMap map[int64][]oldCachedInvite
}{CacheMap: make(map[int64][]oldCachedInvite)}

func BenchmarkOldInviteLogic(b *testing.B) {
	b.ReportAllocs()

	const guildID = 10
	const id = "abcdefgh"

	s := make([]oldCachedInvite, 10)
	for i := 0; i < 10; i++ {
		s[i] = oldCachedInvite{time.Now(), "hgfedcba"}
	}

	for i := int64(0); i < 10000; i++ {
		oldInvitesCache.CacheMap[i] = s
	}
	oldInvitesCache.CacheMap[10001] = []oldCachedInvite{{time.Now(), id}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			oldInvitesCache.RLock()
			for guild, cache := range oldInvitesCache.CacheMap {
				for _, cachedInvite := range cache {
					if cachedInvite.Invite == id {
						oldInvitesCache.RUnlock()
						if guild == guildID {
							return
						}
						return
					}
				}
			}
			oldInvitesCache.RUnlock()
		}()
	}
}

func BenchmarkNewInviteLogic(b *testing.B) {
	b.ReportAllocs()

	const guildID = 10
	const id = "abcdefgh"

	for i := int64(0); i < 100000; i++ {
		invitesCache.set(strconv.FormatInt(i, 10), guildID)
	}
	invitesCache.set(id, guildID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			i, ok := invitesCache.get(id)
			if ok {
				if i.guildID == guildID {
					return
				}

				return
			}
		}()
	}
}
