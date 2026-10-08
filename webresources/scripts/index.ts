import { App, HgNamespace } from "../modules/app";

const app = new App();

const g = globalThis as unknown as { hg: HgNamespace };
g.hg = app.hgNamespace;
