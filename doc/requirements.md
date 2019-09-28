# Requirements

Some preliminary list of requirements for the system.
It's not that this list is approved or agreed-on in some way,
at this point it's just to keep track of things people mention.

Overall we want to improve experience and automation for 3 groups:

* contributors (important special case: sending first patch)
* maintainers
* reviewers

List of requirements:

* Not-email-based (for reasons outlined [here](https://people.kernel.org/monsieuricon/patches-carved-into-developer-sigchains)).
* Compatible with email. At least initially we need bi-directional bridge
  to email with proper threading, etc. Long-term that may be useful as well
  b/c email inbox tend to work as a hub for events happening in different systems.
* User identity. Required for lots of automation (systems doing things on behalf
  of people), tracking/quotas pruposes, being able to describe group maintainership
  policies, etc.
* Distributed/no vendor lock-in.
* Being able to support several types of user interfaces (command line,
  terminal GUI, scripting, web). There are strong cases for all of them.
  So should not be tried with a single UI.
* Support for running offline.
* Easy to setup (especially for local installations and new contributors).
* First-class patch series and patch versions support.
* Subsystem support (what subsystem a change is meant to, mailing list to notify, etc).
* git integration:
  * send change from the current branch
  * create local commits from changes
  * add Reviewed/Acked-by tags to commits
* Automatically importing all changes (series/versions) to a git tree for browsing/pulling.
* Change status tracking:
  * pending changes
  * changes status
  * assigning reviewers
* Maintainers should be able to fix up code/commit descriptions.
* Integration with CIs/static analysis.
* Bug/issue/security vulnerability tracking.
* Stable backport status tracking.
* Encrypted rooms for work on hardware bugs.
