import { App, HgNamespace } from "../modules/app";

const app = new App();

const g = globalThis as unknown as { hg: HgNamespace | undefined };
g.hg = app.hgNamespace;
