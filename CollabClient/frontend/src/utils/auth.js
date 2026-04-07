const TOKEN_KEY = 'jwt_token'

export function migrateLegacyAuthToken() {
    const legacyToken = localStorage.getItem(TOKEN_KEY)
    if (!legacyToken) {
        return
    }

    if (!sessionStorage.getItem(TOKEN_KEY)) {
        sessionStorage.setItem(TOKEN_KEY, legacyToken)
    }
    localStorage.removeItem(TOKEN_KEY)
}

export function setAuthToken(token) {
    if (!token) {
        clearAuthToken()
        return
    }

    sessionStorage.setItem(TOKEN_KEY, token)
    localStorage.removeItem(TOKEN_KEY)
}

export function getAuthToken() {
    return sessionStorage.getItem(TOKEN_KEY) || ''
}

export function clearAuthToken() {
    sessionStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(TOKEN_KEY)
}
