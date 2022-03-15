import URI from 'urijs'

export function transformURL(url: string, targetDomain: string): string {
    const result = new URI()
    const target = new URI(url)

    if (url.length === 0 || url.startsWith("blob:") || url.startsWith("data:")) {
        return url
    }

    if (target.host() === window.location.host && target.hasQuery("proxyTargetURI")) {
        return url
    }

    if (target.host().length === 0 || target.scheme.length === 0) {
        target.host(targetDomain)
        target.protocol('http')
    }

    result.path('')
    result.protocol('http')
    result.host(window.location.host)
    result.setQuery("proxyTargetURI", target.toString())

    return result.toString()
}

export function handleAttr(
    attr: string, element: HTMLElement,
    callback: (value: string) => string
) {
    const value = element.getAttribute(attr)
    if (!value) return
    const modified = callback(value)
    if (modified !== value) {
        element.setAttribute(attr, modified)
    }
}

export function handleSrcSetAttr(element: HTMLElement, targetDomain: string) {
    handleAttr(
        "srcset", element,
        (value) => {
            const result = []
            const candidates = value.split(',')

            for (const c of candidates) {
                const components = c.trim().split(' ')
                components[0] = transformURL(components[0].trim(), targetDomain)
                result.push(components.join(' '))
            }

            return result.join(',')
        }
    )
}

export function handleArchiveAttr(element: HTMLElement, targetDomain: string) {
    handleAttr(
        "archive", element,
        (value) => {
            const urls = value.split(',')
            for (let i = 0; i < urls.length; i++) {
                urls[i] = transformURL(urls[i].trim(), targetDomain)
            }
            return urls.join(',')
        }
    )
}

export function handleContentAttr(element: HTMLElement, targetDomain: string) {
    handleAttr(
        "content", element,
        (value) => {
            const components = value.split(';')
            if (components.length < 2) return value
            components[1] = transformURL(components[1].trim(), targetDomain)
            return components.join(';')
        }
    )
}
