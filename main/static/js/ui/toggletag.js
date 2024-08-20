function toggletag(tag) { //"history"
    var historyDiv = document.getElementById(tag);
    if (historyDiv.style.display === "none") {
        historyDiv.style.display = "block";
    } else {
        historyDiv.style.display = "none";
    }
}