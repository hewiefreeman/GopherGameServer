window.addEventListener("load", init);
var	download_nav,
	download_btn,
	logo,
	title;

function init() {
	download_nav = document.getElementById("downloadNav");
	download_btn = document.getElementById("downloadBtn");
	logo = document.getElementById("logo");
	title = document.getElementById("title");

	window.addEventListener("resize", resized);
	//
	resize_init();
}