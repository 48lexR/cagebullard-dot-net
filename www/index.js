let user = "";

const postGeneric = (id, author, title) => {
    return `<article>
<h2><a href="/posts?id=${id}">${title}</a></h2>
<p>By: <em>${author}</em></p>
</article>`
}

const fullPost = (id, author, title, body) => {
    return `<section id="header">
<h1>${title}</h1>
<p>By: <em>${author}</em></p>
<p>Share this page: <a href="/posts?id=${id}">www.cagebullard.net/post?id=${id}</a></p>
</section>
<section id="body">
${body}
</section>`
}

const loadPosts = async () => {
    const recentPosts = document.getElementById("recents");
    recentPosts.innerHTML = "";
    try {
        const tmp = await fetch("/api/posts")
        if (!tmp.ok) throw new Error("we had an oopsie.");
        const posts = await tmp.json();
        posts.forEach(post => {
            recentPosts.innerHTML += postGeneric(post.postid, post.author, post.title);
        })
    } catch (e) {
        console.error(e);
    }
}

const getPostsByQueryParams = async (container) => {
    const author = new URLSearchParams(window.location.search).get("author")
    const title = new URLSearchParams(window.location.search).get("title")
    const contains = new URLSearchParams(window.location.search).get("contains")
    try {
        let queryString = "/api/posts";
        if (author || title || contains){
            queryString += "?"

        }
        const res = await fetch(queryString);

        if (res.status === 404){
            container.textContent = "not found bud";
            return;
        }
        if (!res.ok){
            container.textContent = "oops...";
            return;
        }

        const posts = await res.json();
        container.innerHTML = "";
        posts.forEach(elem => {
            container.innerHTML += postGeneric(id, post.author, post.title)
        })
    }catch (e) {
        container.textContent = "Couldn't find it..."
    }
}

const getPost = async () => {
    const id = new URLSearchParams(window.location.search).get("id");
    const container = document.getElementById("post");

    if (!id){
        await getPostsByQueryParams(container);
        return;
    }

    try {
        const res = await fetch(`/api/posts/${encodeURIComponent(id)}`);
        if (res.status === 404){
            container.innerHTML = "<p>Oops! We couldn't find that post.</p>";
            return;
        }
        if (!res.ok) {
            container.innerHTML = "<p>Oops! We couldn't find that post.</p>";
            return;
        }
        const post = await res.json();
        container.innerHTML = "";
        container.innerHTML = fullPost(id, post.author, post.title, post.body);
    } catch (e) {
        container.textContent = `couldn't find the server: ${e}`;
    }
}

const createPost = async () => {
    const title = document.getElementById("posttitle").value;
    const body = document.getElementById("postbody").value;
    if (user === ""){
        console.error("We weren't able to fulfill this request because Author is not set.");
        return;
    }
    if (title === ""){
        console.error("No title!");
        return;
    }

    try {
        const res = await fetch("/api/posts", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                "title": title,
                "author": user,
                "body": body
            })
        })

        if (res.status === 404){
            console.error("error");
            return
        }

        if (!res.ok){
            console.error("error");
            return;
        }

        const data = await res.json();
        console.log(data);
        window.location.replace(`/posts?id=${data.postid}`)
    } catch (e) {
        console.error("Unable to fulfill this request because of a server error: ", e);
    }
}
