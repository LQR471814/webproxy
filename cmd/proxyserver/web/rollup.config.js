import { terser } from 'rollup-plugin-terser'
import typescript from '@rollup/plugin-typescript'
import commonjs from '@rollup/plugin-commonjs'
import resolve from '@rollup/plugin-node-resolve'

const plugins = [
    commonjs(),
    resolve(),
    typescript(),
    terser({format: { comments: false}}),
]

export default [
    {
        input: "src/injector/post.ts",
        output: {
            file: "dist/post.min.js",
            format: 'cjs'
        },
        plugins: plugins,
    },
    {
        input: "src/injector/pre.ts",
        output: {
            file: "dist/pre.min.js",
            format: 'cjs'
        },
        plugins: plugins,
    },
]
