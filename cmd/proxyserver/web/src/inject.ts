import { handleArchiveAttr, handleAttr, handleContentAttr, handleSrcSetAttr, transformURL } from './urls'
import { BufferedObserver } from './utils'

const TARGET_DOMAIN = '${TARGET_DOMAIN}'
const LINK_ATTRIBUTES = [
    "src",
    "href",
    "cite",
    "background",
    "codebase",
    "action",
    "longdesc",
    "profile",
    "usemap",
    "classid",
    "data",
    "formaction",
    "icon",
    "poster",
    "srcset",
    "archive",
    "content",
]

function handleAttribute(attribute: string, element: HTMLElement) {
    //? Handle different url formats
    switch (attribute) {
        case "srcset":
            handleSrcSetAttr(element, TARGET_DOMAIN)
            break
        case "archive":
            handleArchiveAttr(element, TARGET_DOMAIN)
            break
        case "content":
            handleContentAttr(element, TARGET_DOMAIN)
            break
        default:
            handleAttr(
                attribute, element,
                (value) => transformURL(value, TARGET_DOMAIN)
            )
            break
    }
}

// document.body.addEventListener('load', () => {
for (const attr of LINK_ATTRIBUTES) {
    const elements = document.evaluate(
        `//*[@${attr}]`, document, null,
        XPathResult.ORDERED_NODE_SNAPSHOT_TYPE, null
    )
    for (let i = 0; i < elements.snapshotLength; i++) {
        handleAttribute(attr, elements.snapshotItem(i) as HTMLElement)
    }
}

const observer = new BufferedObserver(function (mutationList, _) {
    console.info('Mutated', mutationList)
    for (const mutation of mutationList) {
        const mutated = mutation.target as HTMLElement
        if (!mutation.attributeName) continue

        const value = mutated.getAttribute(mutation.attributeName)
        if (!value) continue

        handleAttribute(mutation.attributeName, mutated)
    }
}, 10)

observer.observe(document, {
    subtree: true,
    attributes: true,
    attributeFilter: LINK_ATTRIBUTES
})
// })
