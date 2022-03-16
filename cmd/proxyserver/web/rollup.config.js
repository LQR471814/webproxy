import { terser } from 'rollup-plugin-terser'
import typescript from '@rollup/plugin-typescript'
import commonjs from '@rollup/plugin-commonjs'
import resolve from '@rollup/plugin-node-resolve'

export default {
    input: "src/inject.ts",
    output: {
        file: "dist/inject.min.js",
        format: 'cjs'
    },
    plugins: [
        commonjs(),
        resolve(),
        typescript(),
        terser()
    ],
}
