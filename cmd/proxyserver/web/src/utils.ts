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
