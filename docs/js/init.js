window.addEventListener("load", init);
var	download_nav,
	download_btn,
	wiki_nav,
	wiki_btn,
	logo,
	title;

function init() {
	download_nav = document.getElementById("downloadNav");
	download_btn = document.getElementById("downloadBtn");
	wiki_nav = document.getElementById("wikiNav");
	wiki_btn = document.getElementById("wikiBtn");
	logo = document.getElementById("logo");
	title = document.getElementById("title");

	window.addEventListener("resize", resized);
	//
	resize_init();
}