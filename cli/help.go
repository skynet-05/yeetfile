package main

const mainHelp = `yeetfile

* Commands
  - Account:
      signup
      login
      logout
  - Files:
      upload
      download

Examples:
yeetfile signup
yeetfile upload -e 10d documents.zip
yeetfile download unique.file.tag
`

const uploadHelp = `yeetfile upload

Args:
-d, --downloads  : Set # of times a file can be downloaded
-e, --expiration : Set the lifetime of the file, using the format
   <value><unit>, where value is a numeric and unit is one of
   the following:
       s (seconds)
       m (minutes)
       h (hours)
       d (days)
   Example:
       2d == 2 days
       3h == 3 hours
       20m == 20 minutes
Examples:
yeetfile upload -d 2 -e 10d documents.zip
yeetfile upload -d 5 -e 2h game.exe
`

const downloadHelp = `yeetfile download
Args: None
Examples:
    yeetfile download https://yeetfile.com/d/unique.file.path
    yeetfile download other.unique.path
`

const signupHelp = `yeetfile signup

Args: None
Examples:
    yeetfile signup
`

const loginHelp = `yeetfile login

Args: None
Examples:
    yeetfile login
`

const logoutHelp = `yeetfile logout

Args: None
Examples:
    yeetfile logout
`
