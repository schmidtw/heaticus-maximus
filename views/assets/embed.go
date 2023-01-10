// SPDX-FileCopyrightText: 2023 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package assets

import (
	"embed"
)

//go:generate rm -rf .tmp/chartjs
//
// Fetch upstream Chart.js blob
//
//go:generate go-getter https://github.com/chartjs/Chart.js/releases/download/v4.1.2/chart.js-4.1.2.tgz?archive=false&checksum=ba785a82b3a142148f8d1cb90de2e014729ced3ea178775f294b1ed23e5cbdc2 .cache/
//go:generate rm -rf .tmp/chartjs
//go:generate mkdir -p .tmp/chartjs
//go:generate tar -xzf .cache/chart.js-4.1.2.tgz -C .tmp/chartjs --strip-components 2 package/dist/chart.umd.js

//
// Fetch upstream Fomantic-UI blob
//
//go:generate go-getter https://github.com/fomantic/Fomantic-UI/archive/refs/tags/2.9.0.tar.gz?archive=false&checksum=82eb197367cec5edd246a695ca9039227249ea7116d494408237f39e24081518&filename=fomantic-2.9.0.tgz .cache/
//go:generate rm -rf .tmp/fomantic
//go:generate mkdir -p .tmp/fomantic
//go:generate tar -xzf .cache/fomantic-2.9.0.tgz -C .tmp/fomantic --strip-components 1

//
// Compose the filesystem so embed.FS can do it's magic.
//
// Start from a clean directory.
//go:generate rm -rf site
//go:generate mkdir -p site/js site/css
//
//go:generate cp    .tmp/chartjs/chart.umd.js            site/js/chart.js
//go:generate cp    .tmp/fomantic/dist/semantic.min.js   site/js/.
//go:generate cp    .tmp/fomantic/dist/semantic.min.css  site/css/.
//go:generate cp -a .tmp/fomantic/dist/themes            site/css/.
//
// Clean up from the tarballs so tree command is still useful.
//go:generate rm -rf .tmp

//go:embed site/css
var SiteCSS embed.FS

//go:embed site/js
var SiteJS embed.FS
