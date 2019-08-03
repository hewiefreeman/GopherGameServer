var download_btn_hidden = false;
var download_btn_switch = 520;

function resized(e) {
	if (document.documentElement.clientWidth <= download_btn_switch && !download_btn_hidden) {
		download_nav.style.display = "inline-block";
		download_btn.style.display = "none";

		download_btn_hidden = true;
	} else if (document.documentElement.clientWidth > download_btn_switch && download_btn_hidden) {
		download_nav.style.display = "none";
		download_btn.style.display = "block";

		download_btn_hidden = false;
	}
}

function resize_init() {
	if (document.documentElement.clientWidth <= download_btn_switch) {
		download_nav.style.display = "inline-block";
		download_btn.style.display = "none";

		download_btn_hidden = true;
	} else if (document.documentElement.clientWidth > download_btn_switch) {
		download_nav.style.display = "none";
		download_btn.style.display = "block";

		download_btn_hidden = false;
	}
}