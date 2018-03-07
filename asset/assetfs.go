package asset

import "github.com/elazarl/go-bindata-assetfs"

var AssetFs = assetfs.AssetFS{
	Asset:     Asset,
	AssetDir:  AssetDir,
	AssetInfo: AssetInfo,
	Prefix:    "swagger-ui",
}
