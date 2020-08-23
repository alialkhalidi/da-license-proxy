## GML: Get Me a License, fronts an appsim server, and gets a DA license for an MYAM user to be used for IDVS getAndMatchData call
### Use:
- go install gml
- bin/gml gmlserverconfig.yml

### TODO:
- implement proper retry logic: refactor SendRequestAndCheckResponse and add retry logic based on response code. TL;DR RecoverLockBox sometimes gets 504 Gateway Timeout in UAT.

