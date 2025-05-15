module nextcloud-sharing-client

go 1.23

require github.com/nextcloud/api-sdk v0.0.0

require github.com/studio-b12/gowebdav v0.10.0 // indirect

replace github.com/nextcloud/api-sdk => ../client-sdks/go

//require github.com/nextcloud/client-sdks v0.0.0

//require github.com/nextcloud/client-sdks/go v0.0.0-20240807154148-02bc7281ee36 // indirect
