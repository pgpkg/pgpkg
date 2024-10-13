-- Test handling of comments in managed objects.
create view gl.account_view as select * from gl.account;

comment on view gl.account_view is 'Test comment';
-- comment on column gl.account_view.team is 'Team comment';