# Copyright 2026 Gentoo Authors
# Distributed under the terms of the GNU General Public License v2

EAPI=8

inherit meson

if [[ ${PV} == 9999 ]]; then
	inherit git-r3
	EGIT_REPO_URI="https://repo.dec05eba.com/gpu-screen-recorder-notification"
else
	SRC_URI="https://dec05eba.com/snapshot/${PN}.git.${PV}.tar.gz -> ${P}.tar.gz"
	S="${WORKDIR}"
	KEYWORDS="~amd64 ~arm64"
fi

DESCRIPTION="Lightweight, low-overhead notification overlay for GPU Screen Recorder"
HOMEPAGE="https://git.dec05eba.com/gpu-screen-recorder-notification/about"
LICENSE="GPL-3"
SLOT="0"

DEPEND="
	dev-libs/wayland
	x11-libs/libXext
	x11-libs/libX11
	media-libs/freetype
	media-libs/libglvnd
"
RDEPEND="
	${DEPEND}
	media-video/gpu-screen-recorder
"
BDEPEND="
	${DEPEND}
	virtual/pkgconfig
	dev-util/wayland-scanner
"

src_configure() {
	local emesonargs=(
		# Force meson to use system dependencies instead of downloading git subprojects
		--wrap-mode=nodownload
	)
	meson_src_configure
}

src_install() {
	meson_src_install
}