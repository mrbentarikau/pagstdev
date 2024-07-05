package reminders

var DBSchemas = []string{`
CREATE TABLE IF NOT EXISTS reminders (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE,
	-- text instead of bigint for legacy compatibility
	user_id TEXT NOT NULL,
	channel_id TEXT NOT NULL,
	guild_id BIGINT NOT NULL,
	message TEXT NOT NULL,
	"when" BIGINT NOT NULL,
	repeat BIGINT NOT NULL,
	app_command BOOLEAN NOT NULL DEFAULT FALSE,
	disable_mention BOOLEAN NOT NULL DEFAULT FALSE
);
`, `
CREATE INDEX IF NOT EXISTS idx_reminders_deleted_at ON reminders(deleted_at);
`, `
-- Previous versions of the reputation module used gorm instead of sqlboiler,
-- which does not add NOT NULL constraints by default. Therefore, ensure the
-- NOT NULL constraints are present in existing tables as well.
-- The first few columns below have always been set since the reminders plugin was
-- added, so barring the presence of invalid entries, we can safely add NOT NULL
-- constraints without error.
ALTER TABLE reminders ALTER COLUMN created_at SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN updated_at SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN user_id SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN channel_id SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN message SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN "when" SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN repeat SET NOT NULL;
`, `
DO $$
BEGIN
-- only run if we haven't added the NOT NULL constraint yet
IF EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='reminders' AND column_name='app_command') THEN
	UPDATE reminders SET app_command = FALSE WHERE app_command IS NULL;
	UPDATE reminders SET disable_mention = FALSE WHERE disable_mention IS NULL;
	ALTER TABLE reminders ALTER COLUMN app_command SET NOT NULL;
	ALTER TABLE reminders ALTER COLUMN disable_mention SET NOT NULL;
END IF;
END $$;
`, `
DO $$
BEGIN
-- The guild_id column is more annoying to deal with. When the reminders plugin
-- was first created, the reminders table did not have a guild_id column -- it
-- was added later, in October 2018 (9f5ef28). So reminders before then could
-- plausibly have guild_id = NULL, meaning directly adding the NOT NULL
-- constraint would fail. But since the maximum offset of a reminder is 1 year,
-- all such reminders have now expired and so we can just delete them before
-- adding the constraint.
-- Only run if we haven't added the NOT NULL constraint yet.
IF EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='reminders' AND column_name='guild_id' AND is_nullable='YES') THEN
	DELETE FROM reminders WHERE guild_id IS NULL;
	ALTER TABLE reminders ALTER COLUMN guild_id SET NOT NULL;
END IF;
END $$;
`, `
ALTER TABLE reminders ALTER COLUMN app_command SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN app_command SET DEFAULT FALSE;
`, `
ALTER TABLE reminders ALTER COLUMN disable_mention SET NOT NULL;
`, `
ALTER TABLE reminders ALTER COLUMN  disable_mention SET DEFAULT FALSE;
`}
