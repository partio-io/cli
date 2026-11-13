package entitydiff

import "testing"

func TestAnalyzeJS(t *testing.T) {
	tests := []struct {
		name  string
		patch string
		want  string
	}{
		{
			name: "added function, class, interface, type alias and enum",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,1 +1,12 @@",
				" import { x } from './x';",
				"+export function load(id: string): Promise<Data> {",
				"+  return fetch(id);",
				"+}",
				"+export class Store<T> extends Base<{ a: T }> implements Saver {",
				"+  save(x: T) {}",
				"+}",
				"+export interface Options extends Base, Other {",
				"+  dryRun: boolean;",
				"+}",
				"+export type Handler<T = unknown> = (x: T) => void;",
				"+enum Color { Red, Green }",
			)...),
			want: "a.ts\n" +
				"  added load: export function load(id: string): Promise<Data>\n" +
				"  added Store: export class Store<T> extends Base<{ a: T }> implements Saver\n" +
				"  added Options: export interface Options extends Base, Other\n" +
				"  added Handler: export type Handler<T = unknown>\n" +
				"  added Color: enum Color\n",
		},
		{
			name: "removed class and function",
			patch: patch(with(fileHeader("a.js"),
				"@@ -1,5 +1,1 @@",
				" 'use strict';",
				"-class Legacy {",
				"-  run() {}",
				"-}",
				"-function gone(a, b) {}",
			)...),
			want: "a.js\n  removed Legacy: class Legacy\n  removed gone: function gone(a, b)\n",
		},
		{
			name: "signature changed",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,4 +1,4 @@",
				"-export async function read(id: string): Promise<void> {",
				"+export async function read(id: string, opts?: Options): Promise<Data> {",
				"   return;",
				" }",
				"-export const enum Mode {",
				"+export enum Mode {",
			)...),
			want: "a.ts\n" +
				"  signature changed read: export async function read(id: string): Promise<void> => export async function read(id: string, opts?: Options): Promise<Data>\n" +
				"  signature changed Mode: export const enum Mode => export enum Mode\n",
		},
		{
			name: "modifiers",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -0,0 +1,6 @@",
				"+export default function main() {}",
				"+export default class App {}",
				"+export abstract class Shape {}",
				"+export declare function ambient(x: number): void;",
				"+export async function* stream() {}",
				"+export default function () {}",
			)...),
			want: "a.ts\n" +
				"  added main: export default function main()\n" +
				"  added App: export default class App\n" +
				"  added Shape: export abstract class Shape\n" +
				"  added ambient: export declare function ambient(x: number): void\n" +
				"  added stream: export async function* stream()\n",
		},
		{
			name: "arrow and function-expression bindings",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -0,0 +1,10 @@",
				"+export const handler = async (req: Request): Promise<Response> => {",
				"+const double = x => x * 2;",
				"+let make = <T,>(x: T): { v: T } => ({ v: x });",
				"+var legacy = function (a) {",
				"+export const typed: Handler = async function named(a: number) {",
				"+const count = 5;",
				"+const config = {",
				"+const sum = (a + b) * 2;",
				"+const pick = cond ? (a) => a : null;",
				"+let later: string;",
			)...),
			want: "a.ts\n" +
				"  added handler: export const handler = async (req: Request): Promise<Response>\n" +
				"  added double: const double = x\n" +
				"  added make: let make = <T>(x: T): { v: T }\n" +
				"  added legacy: var legacy = function (a)\n" +
				"  added typed: export const typed: Handler = async function named(a: number)\n",
		},
		{
			name: "control flow, class methods and object members are not entities",
			patch: patch(with(fileHeader("a.js"),
				"@@ -1,2 +1,17 @@",
				" class Box {",
				"+  run(x) {",
				"+  async load() {",
				"+  static function() {}",
				"+  function helper() {}",
				"+  const inner = () => {};",
				" }",
				"+if (x) {",
				"+for (const f of fs) {",
				"+while (go()) {",
				"+switch (k) {",
				"+export default {",
				"+  method() {},",
				"+  arrow: () => {},",
				"+  fn: function (a) {},",
				"+};",
				"+foo(function bar() {});",
			)...),
			want: "",
		},
		{
			name: "change on a continuation line with the function line as context",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,4 +1,5 @@",
				" export function f(",
				"   a: number,",
				"+  b: string,",
				" ): void {",
				"   return;",
			)...),
			want: "a.ts\n  signature changed f: export function f(a: number): void => export function f(a: number, b: string): void\n",
		},
		{
			name: "multi-line arrow with a continuation-line change",
			patch: patch(with(fileHeader("a.js"),
				"@@ -1,4 +1,4 @@",
				" export const f = async (",
				"-  a,",
				"+  a, b,",
				" ) => {",
				"   return a;",
			)...),
			want: "a.js\n  signature changed f: export const f = async (a) => export const f = async (a, b)\n",
		},
		{
			name: "whitespace-only reflow and body-only edit are not reported",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,3 +1,7 @@",
				"-function f(a: number, b: string): void { // doc",
				"+function f(",
				"+  a: number,",
				"+  b: string,",
				"+): void {",
				"-  return;",
				"+  const g = () => 1;",
				" }",
			)...),
			want: "",
		},
		{
			name: "class header across lines",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,4 +1,4 @@",
				" export class Repository",
				"   extends Base",
				"-  implements Reader",
				"+  implements Reader, Writer",
				" {",
			)...),
			want: "a.ts\n  signature changed Repository: export class Repository extends Base implements Reader => export class Repository extends Base implements Reader, Writer\n",
		},
		{
			name: "return type literals do not end the signature",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,2 +1,2 @@",
				"-function f(): { a: number } {",
				"+function f(): { a: number; b: string } | null {",
				"-function g<T extends { id: string }>(x: T): () => { ok: boolean } {",
				"+function g<T extends { id: string }>(x: T): () => { ok: true } {",
			)...),
			want: "a.ts\n" +
				"  signature changed f: function f(): { a: number } => function f(): { a: number; b: string } | null\n" +
				"  signature changed g: function g<T extends { id: string }>(x: T): () => { ok: boolean } => function g<T extends { id: string }>(x: T): () => { ok: true }\n",
		},
		{
			name: "strings and comments in a signature",
			patch: patch(with(fileHeader("a.js"),
				"@@ -1,1 +1,1 @@",
				"-function f(a = '(', /* old */ b) { // {",
				"+function f(a = ')', /* new */ b) { // {",
			)...),
			want: "a.js\n  signature changed f: function f(a = '(', b) => function f(a = ')', b)\n",
		},
		{
			name: "declarations exported from a namespace",
			patch: patch(with(fileHeader("types.d.ts"),
				"@@ -1,3 +1,4 @@",
				" declare module 'lib' {",
				"-  export function run(): void;",
				"+  export function run(opts: Opts): void;",
				"+  export interface Opts {}",
				" }",
			)...),
			want: "types.d.ts\n" +
				"  signature changed run: export function run(): void => export function run(opts: Opts): void\n" +
				"  added Opts: export interface Opts\n",
		},
		{
			name: "unclosed header stops at the next top-level statement",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -0,0 +1,6 @@",
				"+export function broken(a,",
				"+export function next() {}",
				"+type Pair<",
				"+  A,",
				"+> = [A, A];",
				"+let x: string",
			)...),
			want: "a.ts\n  added next: export function next()\n  added Pair: type Pair<A>\n",
		},
		{
			name: "declaration cut off at a hunk boundary is dropped",
			patch: patch(with(fileHeader("a.ts"),
				"@@ -1,3 +1,2 @@",
				" import x from 'x';",
				"-function g(a) {}",
				"-function f(a) {",
				"+function f(a,",
			)...),
			want: "a.ts\n  removed g: function g(a)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(Analyze(tt.patch)); got != tt.want {
				t.Errorf("Analyze() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestAnalyzeJSExtensions(t *testing.T) {
	tests := []struct {
		path      string
		supported bool
	}{
		{"a.ts", true},
		{"a.tsx", true},
		{"a.mts", true},
		{"a.cts", true},
		{"a.js", true},
		{"a.jsx", true},
		{"a.mjs", true},
		{"a.cjs", true},
		{"index.d.ts", true},
		{"index.d.mts", true},
		{"index.d.cts", true},
		{"vendor/lib.min.js", false},
		{"a.coffee", false},
		{"a.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			p := patch(with(fileHeader(tt.path),
				"@@ -0,0 +1,1 @@",
				"+export function f() {}",
			)...)
			want := ""
			if tt.supported {
				want = tt.path + "\n  added f: export function f()\n"
			}
			if got := render(Analyze(p)); got != want {
				t.Errorf("Analyze() =\n%s\nwant\n%s", got, want)
			}
		})
	}
}
