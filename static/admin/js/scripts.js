// This is the script to manage the sidebar for making the pills active based on the page
anchors = document.getElementsByClassName("clickable")
linksMap = new Map();

window.onload = function() {
    linksMap.clear()    
    Array.from(anchors).forEach((e) => {
        linksMap.set(e.href.split('/').pop().trim().toLowerCase(), e)
    })

    page = location.href.split('/').pop()

    linksMap.forEach((tag, k) => {
        if (k === page) {
            tag.classList.remove('link-dark')
            tag.classList.add('active')
        }
        else {
            tag.classList.add('link-dark')
            tag.classList.remove('active')
        }
    })
};