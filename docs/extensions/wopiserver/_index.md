---
title: WOPI server
weight: 20
geekdocRepo: https://github.com/owncloud/ocis-wopiserver
geekdocEditPath: edit/main/docs/extensions/wopiserver
geekdocFilePath: _index.md
geekdocCollapseSection: true
---

oCIS WOPI server is  a proof of concept extension to open office files in ownCloud Infinite Scale. It uses the [CS3ORG WOPI server](https://github.com/cs3org/wopiserver) and can integrate WOPI compliant online office suites like Collabora or Office Online Server.


## Request flow between services

{{< mermaid class="text-center">}}
sequenceDiagram
    autonumber
    participant User
    participant ownCloud Web
    participant oCIS WOPI server
    participant CS3 WOPI server
    participant oCIS
    participant REVA


    User ->> ownCloud Web: Log in with OpenID connect


    activate ownCloud Web
        Note over ownCloud Web: user session represended by OpenID Connect access token
        User ->> ownCloud Web: open office file

        activate ownCloud Web
            ownCloud Web ->> oCIS: /api/v0/wopi/open [OpenID Connect access token]
            activate oCIS
                oCIS ->> oCIS: mints REVA access token for the user
                oCIS ->> oCIS WOPI server: /api/v0/wopi/open [REVA access token]
                activate oCIS WOPI server

					oCIS WOPI server ->> oCIS WOPI server: mint new REVA user token with specified TTL <br> (default 1h)

					oCIS WOPI server ->> REVA: stat file [REVA access token]
					activate REVA
						REVA -->> oCIS WOPI server: file info
					deactivate REVA
					oCIS WOPI server ->> CS3 WOPI server: get supported file extensions [unauthenticated]
					activate CS3 WOPI server
					CS3 WOPI server -->> oCIS WOPI server: file extensions
					deactivate CS3 WOPI server


                    oCIS WOPI server ->> CS3 WOPI server: /wopi/iop/open [REVA JWT secret, REVA access token]
					activate CS3 WOPI server
					Note right of oCIS WOPI server: no REVA JWT secret should be passed here

                        activate CS3 WOPI server

                            CS3 WOPI server ->> CS3 WOPI server: mints CS3 WOPI server access token <br> embeds the REVA access token of the user
                        deactivate CS3 WOPI server

                        CS3 WOPI server -->> oCIS WOPI server: Collabora URL, CS3 WOPI server token


                    deactivate CS3 WOPI server
                    oCIS WOPI server -->> oCIS: Collabora URL, CS3 WOPI server token
                deactivate oCIS WOPI server
                oCIS -->> ownCloud Web: Collabora URL, CS3 WOPI server token
            deactivate oCIS

            ownCloud Web ->> Collabora: open Collabora in new tab (Collabora URL, CS3 WOPI server token as parameters)
            deactivate ownCloud Web

            activate Collabora

                Collabora ->> CS3 WOPI server: /wopi/files/"<"fileid> [CS3 WOPI server token]

                activate CS3 WOPI server

                    CS3 WOPI server ->> CS3 WOPI server: get REVA access token from inside the CS3 WOPI server token

                    CS3 WOPI server ->> REVA: open file [REVA access token]

                    activate REVA
                        REVA -->> CS3 WOPI server: return file
                    deactivate REVA

                    CS3 WOPI server -->> Collabora: return file
                deactivate CS3 WOPI server


                activate User
                    Collabora -->> User: display the file
                    Note over User: can edit the file

                    User ->> Collabora: stop editing
                deactivate User
            deactivate Collabora

    User ->> ownCloud Web: logout

    deactivate ownCloud Web
{{< /mermaid >}}
