# Limit parser CLI

## Notes
I made a few *executive decisions*. In the README provided it mentioned that if a record came in with the same id 
again for a customer it can be ignored. So I did exactly that, it does not get inserted into the database and does not 
count against load limits. This is easily changed if that is not appropriate.