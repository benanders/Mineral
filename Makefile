
#
#  Makefile
#

# Bundles all assets into a Go file, so that they're included in the final
# executable.
gen:
	@go-bindata -pkg asset -prefix "asset/data" -ignore "\.DS_Store" -o asset/asset.go asset/data/...
