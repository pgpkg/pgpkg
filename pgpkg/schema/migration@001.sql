--
-- pgpkg no longer uses the full path name to track migrations, just the filename.
-- this change allows users to move migration scripts around when reorganising a project,
-- without requiring any special kind of registration or SQL comment mangling.
--
update pgpkg.migration set path = regexp_replace(path, '^.*/', '') where path like '%/%';