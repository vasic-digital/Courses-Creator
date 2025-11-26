import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('electronAPI', {
  selectMarkdownFile: () => ipcRenderer.invoke('select-markdown-file'),
  selectOutputDirectory: () => ipcRenderer.invoke('select-output-directory'),
  readFile: (filePath: string) => ipcRenderer.invoke('read-file', filePath),
  writeFile: (filePath: string, content: string) => ipcRenderer.invoke('write-file', filePath, content),
});