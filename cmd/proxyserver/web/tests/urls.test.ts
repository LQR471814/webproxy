/**
 * @jest-environment jsdom
 */

import { transformURL } from "../src/injector/urls";

const proxyDomain = 'proxy.com:3000'
global.window = Object.create(window)
Object.defineProperty(window, 'location', {
    value: {
        host: proxyDomain
    }
})
expect(window.location.host).toBe(proxyDomain)

test('Test transformURL', () => {
    const vectors = [
        {
            url: 'http://abc.com/file.txt?q=s',
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Fabc.com%2Ffile.txt%3Fq%3Ds`
        },
        {
            url: 'index.js',
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Ftarget.com%2Findex.js`
        },
        {
            url: '/path/12345',
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Ftarget.com%2Fpath%2F12345`
        },
        {
            url: '/path/index.js?q=s',
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Ftarget.com%2Fpath%2Findex.js%3Fq%3Ds`
        },
        {
            url: 'path/index.js?q=s',
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Ftarget.com%2Fpath%2Findex.js%3Fq%3Ds`
        },
        {
            url: '/',
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Ftarget.com%2F`
        },
        {
            url: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Fa.b.c`,
            domain: 'target.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Fa.b.c`
        },
        {
            url: '//en.wikipedia.org/',
            domain: 'wikipedia.org',
            expect: `http://${proxyDomain}/?proxyTargetURI=http%3A%2F%2Fen.wikipedia.org%2F`
        },
        {
            url: ``,
            domain: 'target.com',
            expect: ``
        },
        {
            url: `blob:https://example.org/40a5fb5a-d56d-4a33-b4e2-0acf6a8e5f64`,
            domain: 'target.com',
            expect: `blob:https://example.org/40a5fb5a-d56d-4a33-b4e2-0acf6a8e5f64`
        },
        {
            url: `data:text/plain;base64,dGhpcyBpcyBzb21lIHRlc3QgdGV4dA==`,
            domain: 'target.com',
            expect: `data:text/plain;base64,dGhpcyBpcyBzb21lIHRlc3QgdGV4dA==`
        },
    ]
    for (const v of vectors) {
        expect(transformURL(v.url, v.domain))
            .toBe(v.expect)
    }
})
