# Other developer process support systems

In no particular order.

## [git-appraise](https://github.com/google/git-appraise)

Distributed code review system for Git repos:

* Metadata is checked into the git repository itself as notes.
* All users need to have write access to the repo.
* Has command line interface along the lines of: `git appraise request/list/comment`.
* Has (local) web interface.
* Has bridges to github/phabricator.

## [Gerrit](https://www.gerritcodereview.com)

Code review system:

* Metadata is checked into the git repository itself into separate branches.
* Has handy (local) web interface.
* Has dashboard/plugins/access control.

## [git-lab-porcelain](https://gitlab.com/nhorman/git-lab-porcelain)

git porcelain for working with gitlab via its REST api.

## [Patchwork](https://patchwork.ozlabs.org/project/netdev/list/)

Supports command-line interactions with [git-pw](https://patchwork.readthedocs.io/projects/git-pw/en/latest/usage/).\
Commands include: listing, updating/delegating, downloading/applying patches/series. 

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

## [GitHub](https://github.com/)

## [GitLab](https://about.gitlab.com/)

## [icdiff](https://www.jefftk.com/icdiff)

Colored side-by-side diffs in console.
