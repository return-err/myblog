async function refreshAccessToken() {
    const response = await fetch("/api/auth/refresh", {
        method: "POST"
    });
    if (!response.ok) {
        return false;
    }

    const result = await response.json();

    localStorage.setItem("token", result.access_token);

    return true;
}

async function apiFetch(url, options = {}) {
    let token = localStorage.getItem("token");

    let headers = { ...options.headers }
    if (token) {
        headers["Authorization"] = `Bearer ${token}`;
    }

    console.log("token:", token);
    console.log("headers:", headers);

    let response = await fetch(url, {
        ...options,
        headers: headers
    });

    if (response.status !== 401) {
        return response;
    }

    const refreshed = await refreshAccessToken();
    if (!refreshed) {
        return response;
    }

    token = localStorage.getItem("token");
    if (token) {
        headers["Authorization"] = `Bearer ${token}`
    } else {
        delete headers["Authorization"];
    }

    response = await fetch(url, {
        ...options,
        headers: headers
    });

    return response;
}

async function logout() {
    try {
        const token = localStorage.getItem('token')
        if (!token) {
            console.warn("ログインされていません")
            window.location.href = '/login';
            return;
        }

        const response = await fetch('/api/auth/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();
        if (response.ok) {
            console.log("ログアウトしました:", data.message);
            localStorage.removeItem('token');
            window.location.href = '/login';
        } else {
            console.error("ログアウトできませんでした:", data.error || response.statusText)
        }
    } catch (error) {
        console.error("予期せぬエラーが発生しました:", error)
    }
}