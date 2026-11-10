PHP clients have no stable close method (`close()` is experimental), so samples create the client inside the sample function and don't close it.
