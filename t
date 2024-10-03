[1mdiff --git a/go.mod b/go.mod[m
[1mindex 1ea87c1..3855547 100644[m
[1m--- a/go.mod[m
[1m+++ b/go.mod[m
[36m@@ -33,6 +33,7 @@[m [mrequire ([m
 	github.com/DataDog/zstd v1.5.5 // indirect[m
 	github.com/beorn7/perks v1.0.1 // indirect[m
 	github.com/bgentry/speakeasy v0.1.1-0.20220910012023-760eaf8b6816 // indirect[m
[32m+[m	[32mgithub.com/block-vision/sui-go-sdk v1.0.5 // indirect[m
 	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect[m
 	github.com/cenkalti/backoff/v4 v4.1.3 // indirect[m
 	github.com/cespare/xxhash v1.1.0 // indirect[m
[36m@@ -70,6 +71,9 @@[m [mrequire ([m
 	github.com/go-kit/kit v0.12.0 // indirect[m
 	github.com/go-kit/log v0.2.1 // indirect[m
 	github.com/go-logfmt/logfmt v0.6.0 // indirect[m
[32m+[m	[32mgithub.com/go-playground/locales v0.14.1 // indirect[m
[32m+[m	[32mgithub.com/go-playground/universal-translator v0.18.1 // indirect[m
[32m+[m	[32mgithub.com/go-playground/validator/v10 v10.12.0 // indirect[m
 	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect[m
 	github.com/gogo/googleapis v1.4.1 // indirect[m
 	github.com/gogo/protobuf v1.3.2 // indirect[m
[36m@@ -103,6 +107,7 @@[m [mrequire ([m
 	github.com/klauspost/compress v1.17.9 // indirect[m
 	github.com/kr/pretty v0.3.1 // indirect[m
 	github.com/kr/text v0.2.0 // indirect[m
[32m+[m	[32mgithub.com/leodido/go-urn v1.2.2 // indirect[m
 	github.com/lib/pq v1.10.7 // indirect[m
 	github.com/linxGnu/grocksdb v1.8.14 // indirect[m
 	github.com/magiconair/properties v1.8.7 // indirect[m
[36m@@ -139,6 +144,10 @@[m [mrequire ([m
 	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect[m
 	github.com/tendermint/go-amino v0.16.0 // indirect[m
 	github.com/tidwall/btree v1.7.0 // indirect[m
[32m+[m	[32mgithub.com/tidwall/gjson v1.14.4 // indirect[m
[32m+[m	[32mgithub.com/tidwall/match v1.1.1 // indirect[m
[32m+[m	[32mgithub.com/tidwall/pretty v1.2.0 // indirect[m
[32m+[m	[32mgithub.com/tyler-smith/go-bip39 v1.1.0 // indirect[m
 	github.com/zondax/hid v0.9.2 // indirect[m
 	github.com/zondax/ledger-go v0.14.3 // indirect[m
 	go.etcd.io/bbolt v1.3.10 // indirect[m
[36m@@ -149,6 +158,7 @@[m [mrequire ([m
 	golang.org/x/sys v0.23.0 // indirect[m
 	golang.org/x/term v0.23.0 // indirect[m
 	golang.org/x/text v0.17.0 // indirect[m
[32m+[m	[32mgolang.org/x/time v0.5.0 // indirect[m
 	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect[m
 	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect[m
 	google.golang.org/genproto/googleapis/rpc v0.0.0-20240709173604-40e1e62336c5 // indirect[m
[1mdiff --git a/go.sum b/go.sum[m
[1mindex 0720a19..af034a8 100644[m
[1m--- a/go.sum[m
[1m+++ b/go.sum[m
[36m@@ -73,6 +73,8 @@[m [mgithub.com/bgentry/speakeasy v0.1.1-0.20220910012023-760eaf8b6816 h1:41iFGWnSlI2[m
 github.com/bgentry/speakeasy v0.1.1-0.20220910012023-760eaf8b6816/go.mod h1:+zsyZBPWlz7T6j88CTgSN5bM796AkVf0kBD4zp0CCIs=[m
 github.com/bits-and-blooms/bitset v1.8.0 h1:FD+XqgOZDUxxZ8hzoBFuV9+cGWY9CslN6d5MS5JVb4c=[m
 github.com/bits-and-blooms/bitset v1.8.0/go.mod h1:7hO7Gc7Pp1vODcmWvKMRA9BNmbv6a/7QIWpPxHddWR8=[m
[32m+[m[32mgithub.com/block-vision/sui-go-sdk v1.0.5 h1:zM9gJOksgFQkIqJuCi/W4ytwcQYVsOA5AGgbk5WnONc=[m
[32m+[m[32mgithub.com/block-vision/sui-go-sdk v1.0.5/go.mod h1:5a7Ubw+dC2LjdsL+zWMEplSA622MPmTp8uGL5sl1rRY=[m
 github.com/btcsuite/btcd/btcec/v2 v2.3.4 h1:3EJjcN70HCu/mwqlUsGK8GcNVyLVxFDlWurTXGPFfiQ=[m
 github.com/btcsuite/btcd/btcec/v2 v2.3.4/go.mod h1:zYzJ8etWJQIv1Ogk7OzpWjowwOdXY1W/17j2MW85J04=[m
 github.com/btcsuite/btcd/btcutil v1.1.6 h1:zFL2+c3Lb9gEgqKNzowKUPQNb8jV7v5Oaodi/AYFd6c=[m
[36m@@ -256,12 +258,18 @@[m [mgithub.com/go-playground/assert/v2 v2.0.1/go.mod h1:VDjEfimB/XKnb+ZQfWdccd7VUvSc[m
 github.com/go-playground/locales v0.13.0/go.mod h1:taPMhCMXrRLJO55olJkUXHZBHCxTMfnGwq/HNwmWNS8=[m
 github.com/go-playground/locales v0.14.0 h1:u50s323jtVGugKlcYeyzC0etD1HifMjqmJqb8WugfUU=[m
 github.com/go-playground/locales v0.14.0/go.