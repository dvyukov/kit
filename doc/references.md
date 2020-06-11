# Other developer process support systems

In no particular order.

## [git-appraise](https://github.com/google/git-appraise)

Distributed code review system for Git repos:

* Metadata is checked into the git repository itself as notes.
* All users need to have write access to the repo.
* Has command line interface along the lines of: `git appraise request/list/comment`.
* Has (local) web interface.
* Has bridges to github/phabricator.

[Metadata schema](https://github.com/google/git-appraise/tree/master/schema).

## [Gerrit](https://www.gerritcodereview.com)

Code review system:

* Metadata is checked into the git repository itself into separate branches.
* Has handy (local) web interface.
* Has dashboard/plugins/access control.

Information about [metadata schema](https://lore.kernel.org/workflows/87sgn0zr09.fsf@iris.silentflame.com/T/#m3db87b43cf5e581ba4d3a7fd5f1fbff5aea3546a).

## [gertty](https://opendev.org/ttygroup/gertty)

Gertty is a console-based interface to the Gerrit Code Review system.

As compared to the web interface, the main advantages are:

* Workflow -- the interface is designed to support a workflow similar to reading network news or mail. In particular, it is designed to deal with a large number of review requests across a large number of projects.
* Offline Use -- Gertty syncs information about changes in subscribed projects to a local database and local git repos. All review operations are performed against that database and then synced back to Gerrit.
* Speed -- user actions modify locally cached content and need not wait for server interaction.
* Convenience -- because Gertty downloads all changes to local git repos, a single command instructs it to checkout a change into that repo for detailed examination or testing of larger changes.

## [git-lab-porcelain](https://gitlab.com/nhorman/git-lab-porcelain)

git porcelain for working with gitlab via its REST api.

## [Patchwork](https://patchwork.ozlabs.org/project/netdev/list/)

Supports command-line interactions with [git-pw](https://patchwork.readthedocs.io/projects/git-pw/en/latest/usage/).\
Commands include: listing, updating/delegating, downloading/applying patches/series. 

Has read-only API that among other things exposes [information about patches](https://patchwork.ozlabs.org/api/patches/?order=-id).

[Source](https://github.com/getpatchwork/patchwork), [email parser](https://github.com/getpatchwork/patchwork/blob/master/patchwork/parser.py), [parsing tests](https://github.com/getpatchwork/patchwork/tree/master/patchwork/tests/mail).

## [Patchew](https://patchew.org/QEMU/)

Patchwork-fork.

## [public-inbox](https://public-inbox.org/README)

Mailinst list archive system, but has some code reviewing support
(click on links [here](https://public-inbox.org/git/20160711210243.GA1604@whir/)). 

## [Phabricator](https://www.phacility.com/phabricator/)

## [sourcehut](https://sourcehut.org/)

Patchwork-like [dashboard](https://lists.sr.ht/~sircmpwn/sr.ht-dev),
review [example](https://lists.sr.ht/~sircmpwn/ctools/patches/8134).

## [Iron](https://blog.janestreet.com/putting-the-i-back-in-ide-towards-a-github-explorer/)

* Integrated into text editor.
* Comments are committed literally as comments in the code.
* [Video presentation](https://blog.janestreet.com/jane-street-tech-talk-how-jane-street-does-code-review/)

## [Fossil](https://www.fossil-scm.org/home/doc/trunk/www/index.wiki)

I did not really yet understand what it is. But it seems to come with own VCS which is not git.

Fossil is a simple, high-reliability, distributed software configuration management system with these advanced features: 
Integrated Bug Tracking, Wiki, Forum, and Technotes; Built-in Web Interface...

## [GitGitGadget](https://gitgitgadget.github.io/)

Transforms your Pull Request on Github into plain text patch email on a mailing list.

## [email2git](https://github.com/alexcourouble/email2git)

Matching Commits with Their Mailing List Discussions.

## [git-codereview](https://godoc.org/golang.org/x/review/git-codereview)

A set of git commands for code reviews (wrapping gerrit in that implementation) used by Go project.

## [GitHub](https://github.com/)

## [GitLab](https://about.gitlab.com/)

## [icdiff](https://www.jefftk.com/icdiff)

Colored side-by-side diffs in console.

## [Using GIT as a storage]

Also see list at [Using GIT as a storage](https://github.com/ligurio/testres/wiki/Using-GIT-as-a-storage#who-else-supports-git-as-a-storage) and [UI](https://github.com/ligurio/testres/wiki/UI).

