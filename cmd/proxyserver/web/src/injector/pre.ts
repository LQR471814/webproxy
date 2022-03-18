import URI from "urijs"

const FULL_TARGET = '${FULL_TARGET}'
const TARGET_DOMAIN = '${TARGET_DOMAIN}'

window.history.pushState({}, "", FULL_TARGET)

const url = new URI(window.location.toString())
const target = url.query(true)["proxyTargetURI"]
url.removeQuery("proxyTargetURI")
if (target && (url.path() !== "/" || Object.keys(url.query(true)).length > 0)) {
    const targetURL = new URI(decodeURIComponent(target))
        .path(url.path())
        .query(url.query())
    const proxyURL = new URI(window.location.toString())
        .path("/")
        .query(`?proxyTargetURI=${encodeURIComponent(targetURL.toString())}`)
    window.location.href = proxyURL.toString()
}
