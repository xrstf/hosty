Hosty - File Hosting for Minimalists
====================================

Hosty is a tiny Go application to host arbitrary files on your own webserver.
It's meant to be as small as possible and hence offers only the following:

* Users configured in the config file can upload files. Optionally, they can
  also paste code (à la pastebin) with simple syntax highlighting based on
  Pygments.
* Each file gets a randomized URL. Everyone who has the URL can access the
  content.
* Files are simply stored in the filesystem; a minimal SQLite3 database keeps
  track of them.
* Files can destruct themselves on first access and expire after a certain
  amount of time.
* Files can be either public (accessible for everyone who has a link to them),
  internal (only accessible for logged-in users) or private (only for the
  owner).
* Content can be deleted by their respective owners.
* Login via OAuth (Google, GitHub) is possible, but you have to create your own
  apps to do so.

That's it. There is no admin UI, no commenting, no tagging, no folders, no
social media crap or other external resources.

Requirements
------------

* Go 1.20
* On Windows, you need to be able to build [go-sqlite3](https://github.com/mattn/go-sqlite3).

For working on the frontend UI, you will also need Grunt, Bower and npm. Sorry.

Installation
------------

* ``go install go.xrstf.de/hosty`` -- please use ``make`` to make (sic) sure
  that the version stamp is correct.
* Copy the ``config.yaml.dist``, name it as you like (e.g. ``config.yaml``) and
  adjust it to your needs.
* Run hosty via ``./hosty serve config.yaml``.

If you want, you can use the ``resources/systemd/hosty.service`` as an example
for a systemd service.

Screenshots
-----------

### Login screen

![Login Screen](https://h.xrstf.de/r/PGEpXOMTeLjhnQmrTBeThjSVjZVpgMjCvuRBbKaekm/hosty-login.png)

### File Upload View

![File Upload View](https://h.xrstf.de/r/PEzXhNNHgQCXwyrgOuZTLxEUddjQWxcIRWlBMsQWuD/hosty-index.png)

### Mobile View

![Mobile View](https://h.xrstf.de/r/ITeJyfdWXZKnfsTCNfpEihInQzJnqeFRQfqOQorzEG/hosty-mobile.png)

### Image View

![Image View](https://h.xrstf.de/r/MeVpSQXspEWcSxxCVHlHVcwGxqoTZZnZYSNzUBYXFb/hosty-image.png)

### Text View

![Text View](https://h.xrstf.de/r/ZAbkAPInDWAQRUrKnCFpZjLhAGGhkxGqqpoDPPwwWK/hosty-text.png)

### File View

![File View](https://h.xrstf.de/r/BuHLIbyCNMEbbOgDgocQAwigxWIVuOPGSYhmnkPUxG/hosty-file.png)

License
-------

Hosty is licensed under the MIT license.
