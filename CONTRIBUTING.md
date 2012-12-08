Filling issues and contributing to nut tool
===========================================

First of all, thank you for your interest in making packaging better! There are number of ways you can help:

* reporting bugs;
* proposing features;
* contributing code bug fixes and new features;
* contributing documentation fixes (there is probably a ton of grammar errors :/) and improvements.

The following sections describes those scenarios. Golden rule: communicate first, code alter.

Reporting bugs
--------------

1. Make sure bug is reproducible with latest released version: `go get -u github.com/AlekSi/nut/nut`.
2. Search for [existing bug report](https://github.com/AlekSi/nut/issues).
3. Create a new issue if needed. Please do not assign any label.
4. Include output of:

		(cd $GOPATH/src/github.com/AlekSi/nut && git describe --tags)
		go env

5. Include any other information you think may help.

Proposing features
------------------

Please add your comments to [existing feature requests](https://github.com/AlekSi/nut/issues?labels=feature), but do not create new without proposing them in [mailing list](https://groups.google.com/group/gonuts-io) first.

Contributing changes
--------------------

1. Read all previous sections first.
2. Nut tool uses [Git Flow](http://nvie.com/posts/a-successful-git-branching-model/). Make sure you are starting with branch `develop` for new feature and `master` for bug fix.
3. You can make small changes right in the web interface. Spot a typo? Fix it! :)
4. For bigger changes make a fork on GitHub and use `git flow feature start` or `git flow hotfix start`.
5. Make your changes. Do not change version.
6. Publish your `feature` or `hotfix` branch.
7. Make a pull request.
