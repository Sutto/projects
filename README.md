# Projects

Projects is a simple command line tool that discovers git repositories under
a top level directory. I use it with two bash aliases, `lsp` and `cdp` - see
[here for the aliases](https://github.com/Sutto/dot-files/blob/master/home/.bash/rc-ext/02_cdp.sh)
and [here for the completion](https://github.com/Sutto/dot-files/blob/master/home/.bash/profile-ext/completions/projects.sh).

By default, the command line uses `~/Code` as the code directory and caches the items in `~/.cached-projects-list` - but this can
be changed using the `PROJECTS_CODE_PATH` and `PROJECTS_CACHE_PATH` environment variables respectively.

The command can be used as such:

* `projects ls` - List all projects, regenerating the cache if old than five minutes.
* `projects regenerate` - Force regenerating the cache.
* `projects path {item}` - prints the full path for the specified item.

Note that this currently stops recursing once it finds the top level `.git` directory, to avoid hitting
submodules.

This is a partial replacement for the naive ruby version [located here](https://github.com/Sutto/dot-files/blob/master/home/bin/projects).

Released under a Standard MIT license.