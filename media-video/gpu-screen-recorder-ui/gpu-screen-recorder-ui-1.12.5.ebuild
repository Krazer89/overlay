# Copyright 2026 Gentoo Authors
# Distributed under the terms of the GNU General Public License v2

EAPI=8

inherit meson

if [[ ${PV} == 9999 ]]; then
	inherit git-r3
	EGIT_REPO_URI="https://repo.dec05eba.com/gpu-screen-recorder-ui"
else
	SRC_URI="https://dec05eba.com/snapshot/${PN}.git.${PV}.tar.gz -> ${P}.tar.gz"
	S="${WORKDIR}"
	KEYWORDS="~amd64 ~arm64"
fi

DESCRIPTION="A standalone fullscreen overlay UI for GPU Screen Recorder"
HOMEPAGE="https://git.dec05eba.com/gpu-screen-recorder-ui/about"
LICENSE="GPL-3"
SLOT="0"
IUSE="+capabilities +desktop-files"

DEPEND="
	dev-libs/dbus-glib
	dev-libs/wayland
	media-libs/freetype
	media-libs/libglvnd
	media-libs/libpulse
	sys-apps/dbus
	x11-libs/libX11
	x11-libs/libXcomposite
	x11-libs/libXcursor
	x11-libs/libXext
	x11-libs/libXfixes
	x11-libs/libXi
	x11-libs/libXrandr
	x11-libs/libdrm
"
RDEPEND="
	${DEPEND}
	media-video/gpu-screen-recorder
	media-video/gpu-screen-recorder-notification
"
BDEPEND="
	${DEPEND}
	virtual/pkgconfig
	dev-util/wayland-scanner
	x11-misc/gtk-update-icon-cache
"

src_configure() {
	local emesonargs=(
		# Block internal subproject network downloads
		--wrap-mode=nodownload
		$(meson_use capabilities)
		$(meson_use desktop-files)
	)
	meson_src_configure
}

src_install() {
	meson_src_install
}