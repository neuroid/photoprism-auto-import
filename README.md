photoprism-auto-import
======================

[PhotoPrism][1] import helper.

While PhotoPrism offers a [built-in importer][2], it's prone to importing
partially uploaded files.

To overcome this problem, the import helper can be used to watch the
[import folder][3] for changes and only trigger an import after a period of
filesystem inactivity.

[1]: https://github.com/photoprism/photoprism
[2]: https://docs.photoprism.app/user-guide/library/import/#automatic-import
[3]: https://docs.photoprism.app/user-guide/backups/folders/#import

Usage
-----

    Usage: photoprism-auto-import [options] PHOTOPRISM_IMPORT_PATH
      -debug
        	Log filesystem events
      -delay duration
        	How soon after the last filesystem event to trigger the import (default 10s)
      -move
        	Tell PhotoPrism to remove imported files
      -url string
        	PhotoPrism API URL (default "http://127.0.0.1:2342/api/v1/")

    The PHOTOPRISM_APP_PASSWORD environment variable must be set to an app-specific password.
    See https://docs.photoprism.app/user-guide/users/client-credentials/#app-passwords for details.

Example
-------

    export PHOTOPRISM_APP_PASSWORD=xxxxxx-xxxxxx-xxxxxx-xxxxxx
    photoprism-auto-import --move /var/lib/photoprism/import

The above will watch `/var/lib/photoprism/import` for changes and trigger an
import after 10 seconds of filesystem inactivity.

Installation
------------

    go install github.com/neuroid/photoprism-auto-import@latest
