export class BufferedObserver {
    _observer: MutationObserver
    _mutations: MutationRecord[]
    _buffer: number
    _flushTimeout?: number

    callback: MutationCallback
    bufferSize: number

    constructor(callback: MutationCallback, bufferSize: number) {
        this._buffer = 0
        this._mutations = []

        this.bufferSize = bufferSize
        this.callback = callback

        this._observer = new MutationObserver((mutateList) => {
            this._mutations.push(...mutateList)
            this._buffer++
            if (this._buffer >= bufferSize) {
                console.info("Flushed by buffer", this._mutations)
                this._flush()
                return
            }
            if (!this._flushTimeout) {
                this._flushTimeout = window.setTimeout(() => {
                    console.info("Flushed by timeout", this._mutations)
                    this._flush()
                }, 2000)
            }
        })
    }

    _flush = () => {
        this.callback(this._mutations, this._observer)
        this._buffer = 0
        this._mutations = []
        this._flushTimeout = undefined
    }

    observe = (target: Node, options?: MutationObserverInit | undefined) => {
        this._observer.observe(target, options)
    }
}

export type ProxyHooks = {
    get?: (target: any, name: string | symbol) => any
    set?: (target: any, name: string | symbol, value: any) => any
}

export function OverrideProxy(target: any, hooks: ProxyHooks) {
    const descriptors: { [key: string]: PropertyDescriptor } = {}
    for (const prop of Object.getOwnPropertyNames(target)) {
        const description = Object.getOwnPropertyDescriptor(target, prop)
        if (!description) continue
        descriptors[prop] = {
            configurable: description.configurable,
            enumerable: description.enumerable,
            get: () => {
                const value = hooks?.get?.(target, prop) ?? Reflect.get(target, prop)
                if (typeof value !== 'function') return value
                return value.bind(target)
            },
            set: (value) => {
                if (description.writable === false) return false
                const val = hooks?.set?.(target, prop, value)
                if (val === null) return false
                return Reflect.set(target, prop, val ?? value)
            },
        }
    }
    return Object.create({}, descriptors)
}
