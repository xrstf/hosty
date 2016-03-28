Hosty - File Hosting for Minimalists
====================================

Hosty is a tiny Go application to host arbitrary files on your own webserver.
It's meant to be as small as possible and hence offers only the following:

* Users configured in the config file can upload files. Optionally, they can
  also paste code (à la pastebin) with simple syntax highlighting based on
  Pygments.
* Each file gets a randomized URL. Everyone who has the URL can access the
  content.
* Content can be deleted by logged-in users.
* Login via OAuth (Google, GitHub) is possible, but you have to create your own
  apps to do so.

That's it. There is no admin UI, no commenting, no tagging, no folders, no
social media crap or other external resources.

Screenshots
-----------

Login screen
^^^^^^^^^^^^

![Login Screen](https://h.xrstf.de/f/rrieyu2t8r9yfxic0wxp/raw)

File Upload View
^^^^^^^^^^^^^^^^

![File Upload View](https://h.xrstf.de/f/2fuoc5ib65xqdlnhlyha/raw)

Mobile View
^^^^^^^^^^^

![Mobile View](https://h.xrstf.de/f/7lyaghbie1x3eyimz246/raw)

Image View
^^^^^^^^^^

![Image View](https://h.xrstf.de/f/m1oieblh3vbuomzecbsl/raw)

Text View
^^^^^^^^^

![Text View](https://h.xrstf.de/f/7o98vlmqbmel375l1dpf/raw)

File View
^^^^^^^^^

![File View](https://h.xrstf.de/f/119dzazb10d7ht3uk8mz/raw)

Requirements
------------

* Go 1.5
* On Windows, you need to be able to build go-sqlite3.

Files are simply stored in the filesystem; a minimal SQLite3 database keeps track
of them.

License
-------

Hosty is licensed under the MIT license.
