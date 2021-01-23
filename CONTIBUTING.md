# Contributing

Netns exporter is open-source and very open to contributions.

If you're part of a corporation with an NDA, and you may require
updating the license. See Updating Copyright below

## Submitting issues

Issues are contributions in a way so don't hesitate to submit reports on
the [official bugtracker].

Provide as much informations as possible to specify the issues:

-   the netns_exporter version used
-   a stacktrace
-   installed applications list
-   a code sample to reproduce the issue
-   logs
-   information about your system
-   ...

## Submitting patches (bugfix, features, ...)

If you want to contribute some code:

1.  Fork the [official Netns exporter repository](https://github.com/velp/netns-exporter)
2.  Ensure an issue is opened for your feature or bug, if not - do it!
3.  Create a branch with an explicit name (like `my-new-feature` or
    `issue-XX`)
4.  Do your work in it
5.  Commit your changes. Ensure the commit message includes the issue.
    Also, if contributing from a corporation, be sure to add a comment
    with the Copyright information
6.  Rebase it on the master branch from the official repository (cleanup
    your history by performing an interactive rebase)
7.  Ddd your change to the CHANGELOG.md file
8.  Submit your pull-request
9.  2 Maintainers should review the code for bugfix and features. 1
    maintainer for minor changes (such as docs)
10. After review, a maintainer a will merge the PR.

There are some rules to follow:

-   your contribution should be documented (if needed)
-   your contribution should be tested and the test suite should pass
    successfully
-   your code should be mostly golangci-lint checks compatible (run `make lint`)

You need to install some dependencies to develop on jexia-cli:

    $ make install-test

To ensure everything is fine before submission, use `make test` and `make lint`
locally. It will run the basic checks for your code.

    $ make test
    $ make lint
