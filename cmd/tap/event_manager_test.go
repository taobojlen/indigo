package main

import (
	"testing"

	"github.com/bluesky-social/indigo/cmd/tap/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRecordEvents_SkipsRecordTrackingWhenDisabled(t *testing.T) {
	te := newTestEnv(t, testEnvOpts{
		outboxMode:          OutboxModeFireAndForget,
		disableTrackRecords: true,
	})

	te.pushRecordEvents("did:example:user1", 5, true)

	if te.db.Migrator().HasTable(&models.RepoRecord{}) {
		var recordCount int64
		require.NoError(t, te.db.Model(&models.RepoRecord{}).Count(&recordCount).Error)
		assert.Equal(t, int64(0), recordCount, "no RepoRecord rows should be written when tracking is disabled")
	}

	var outboxCount int64
	require.NoError(t, te.db.Model(&models.OutboxBuffer{}).Count(&outboxCount).Error)
	assert.Equal(t, int64(5), outboxCount, "outbox events should still be written")
}

func TestAddRecordEvents_TracksRecordsByDefault(t *testing.T) {
	te := newTestEnv(t, testEnvOpts{
		outboxMode: OutboxModeFireAndForget,
	})

	te.pushRecordEvents("did:example:user1", 5, true)

	var recordCount int64
	require.NoError(t, te.db.Model(&models.RepoRecord{}).Count(&recordCount).Error)
	assert.Equal(t, int64(5), recordCount, "RepoRecord rows should be written by default")
}

func TestSetupDatabase_SkipsRepoRecordMigrationWhenDisabled(t *testing.T) {
	db, err := SetupDatabase("sqlite://file::memory:?cache=shared", 1, false)
	require.NoError(t, err)
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})

	assert.False(t, db.Migrator().HasTable(&models.RepoRecord{}), "repo_records table should not be migrated when tracking is disabled")
	assert.True(t, db.Migrator().HasTable(&models.Repo{}), "repos table should still be migrated")
	assert.True(t, db.Migrator().HasTable(&models.OutboxBuffer{}), "outbox_buffers table should still be migrated")
}
