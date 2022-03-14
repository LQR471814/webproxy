/**
 * @jest-environment jsdom
 */

import { transformURL } from "../src/urls";

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
            url: `http://${proxyDomain}/?proxyTargetURI=https%3A%2F%2Fstackoverflow.com%2Fquestions%2F51040703%2Fwhat-return-type-should-be-used-for-settimeout-in-typescript`,
            domain: 'stackoverflow.com',
            expect: `http://${proxyDomain}/?proxyTargetURI=https%3A%2F%2Fstackoverflow.com%2Fquestions%2F51040703%2Fwhat-return-type-should-be-used-for-settimeout-in-typescript`
        }
    ]
    for (const v of vectors) {
        expect(transformURL(v.url, v.domain))
            .toBe(v.expect)
    }
})
