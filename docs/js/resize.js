var	download_btn_hidden = false,
	download_btn_switch = 520,
	logo_resize = false,
	logo_switch = 350;

function resized(e) {
	// Resize and replace download button
	if (document.documentElement.clientWidth <= download_btn_switch && !download_btn_hidden) {
		download_nav.style.display = "inline-block";
		download_btn.style.display = "none";

		download_btn_hidden = true;
	} else if (document.documentElement.clientWidth > download_btn_switch && download_btn_hidden) {
		download_nav.style.display = "none";
		download_btn.style.display = "block";

		download_btn_hidden = false;
	}

	// Resize logo
	if (document.documentElement.clientWidth <= download_btn_switch) {
		logo_resize = true;
		logo.style.width = (document.documentElement.clientWidth/2)+"px";
		logo.style.height = (document.documentElement.clientWidth/2)+"px";
		title.style.left = ((document.documentElement.clientWidth/2)-30)+"px";
	} else if (document.documentElement.clientWidth > download_btn_switch && logo_resize) {
		logo.style.width = "250px";
		logo.style.height = "250px";
		title.style.left = "220px";
		logo_resize = false;
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

	// Resize logo
	if (document.documentElement.clientWidth <= download_btn_switch) {
		logo_resize = true;
		logo.style.width = (document.documentElement.clientWidth/2)+"px";
		logo.style.height = (document.documentElement.clientWidth/2)+"px";
		title.style.left = ((document.documentElement.clientWidth/2)-30)+"px";
	} else if (document.documentElement.clientWidth > download_btn_switch) {
		logo.style.width = "250px";
		logo.style.height = "250px";
		title.style.left = "220px";
		logo_resize = false;
	}
}