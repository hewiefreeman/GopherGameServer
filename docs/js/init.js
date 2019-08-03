window.addEventListener("load", init);
var	download_nav,
	download_btn;

function init() {
	download_nav = document.getElementById("downloadNav");
	download_btn = document.getElementById("downloadBtn");

	window.addEventListener("resize", resized);
	//
	resize_init();
}