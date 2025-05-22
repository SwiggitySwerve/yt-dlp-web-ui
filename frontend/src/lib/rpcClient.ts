import { Observable } from 'rxjs'
import type { DLMetadata, LiveStreamProgress, RPCRequest, RPCResponse, RPCResult } from '../types'

import { WebSocketSubject, webSocket } from 'rxjs/webSocket'

type DownloadRequestArgs = {
  url: string,
  args: string, // String of CLI style args
  pathOverride?: string,
  renameTo?: string,
  playlist?: boolean,
  channel_folder?: string; 
  preferred_formats?: string[];  // New
  preferred_qualities?: string[]; // New
}

export class RPCClient {
  private seq: number
  private httpEndpoint: string
  private readonly _socket$: WebSocketSubject<any>
  private readonly token?: string

  constructor(httpEndpoint: string, webSocketEndpoint: string, token?: string) {
    this.seq = 0
    this.httpEndpoint = httpEndpoint
    this._socket$ = webSocket<any>({
      url: token ? `${webSocketEndpoint}?token=${token}` : webSocketEndpoint
    })
    this.token = token
  }

  public get socket$(): Observable<RPCResponse<RPCResult[]>> {
    return this._socket$
  }

  private incrementSeq() {
    return String(this.seq++)
  }

  private send(req: RPCRequest) {
    this._socket$.next({
      ...req,
      id: this.incrementSeq(),
    })
  }

  private argsSanitizer(args: string): string[] {
    const splitOnlyWhitespaces = /[^\s"']+|"([^"]*)"|'([^']*)'/gm

    return args
      .match(splitOnlyWhitespaces)
      ?.map(a => a.trim())
      .filter(Boolean) ?? []
  }

  private async sendHTTP<T>(req: RPCRequest) {
    const res = await fetch(this.httpEndpoint, {
      method: 'POST',
      headers: {
        'X-Authentication': this.token ?? ''
      },
      body: JSON.stringify({
        ...req,
        id: this.incrementSeq(),
      })
    })
    const data: RPCResponse<T> = await res.json()

    return data
  }

  public download(req: DownloadRequestArgs) {
    if (!req.url) {
      return
    }

    const rename = req.args.includes('-o')
      ? req.args
        .substring(req.args.indexOf('-o'))
        .replaceAll("'", '')
        .replaceAll('"', '')
        .split('-o')
        .map(s => s.trim())
        .join('')
        .split(' ')
        .at(0) ?? ''
      : ''

    const sanitizedArgs = this.argsSanitizer(
      req.args
        .replace('-o', '')
        .replace(rename, '')
    )

    if (req.playlist) {
      const playlistParamsPayload: any = {
        URL: req.url,
        Params: sanitizedArgs,
        Path: req.pathOverride,
        Rename: req.renameTo || rename,
      };
      if (req.channel_folder) {
        playlistParamsPayload.ChannelFolder = req.channel_folder;
      }
      if (req.preferred_formats) { // New
        playlistParamsPayload.PreferredFormats = req.preferred_formats;
      }
      if (req.preferred_qualities) { // New
        playlistParamsPayload.PreferredQualities = req.preferred_qualities;
      }
      return this.sendHTTP({
        method: 'Service.ExecPlaylist',
        params: [playlistParamsPayload]
      });
    }
    
    const execParamsPayload: any = {
      URL: req.url.split('?list').at(0)!,
      Params: sanitizedArgs,
      Path: req.pathOverride,
      Rename: req.renameTo || rename,
    };
    if (req.channel_folder) {
      execParamsPayload.ChannelFolder = req.channel_folder;
    }
    if (req.preferred_formats) { // New
        execParamsPayload.PreferredFormats = req.preferred_formats;
    }
    if (req.preferred_qualities) { // New
        execParamsPayload.PreferredQualities = req.preferred_qualities;
    }
    this.sendHTTP({
      method: 'Service.Exec',
      params: [execParamsPayload]
    });
  }

  public formats(url: string) {
    if (url) {
      return this.sendHTTP<DLMetadata>({
        method: 'Service.Formats',
        params: [{
          URL: url.split('?list').at(0)!,
        }]
      })
    }
  }

  public running() {
    this.send({
      method: 'Service.Running',
      params: [],
    })
  }

  public kill(id: string) {
    return this.sendHTTP({
      method: 'Service.Kill',
      params: [id],
    })
  }

  public clear(id: string) {
    return this.sendHTTP({
      method: 'Service.Clear',
      params: [id],
    })
  }

  public killAll() {
    return this.sendHTTP({
      method: 'Service.KillAll',
      params: [],
    })
  }

  public freeSpace() {
    return this.sendHTTP<number>({
      method: 'Service.FreeSpace',
      params: [],
    })
  }

  public directoryTree() {
    return this.sendHTTP<string[]>({
      method: 'Service.DirectoryTree',
      params: [],
    })
  }

  public execLivestream(url: string) {
    return this.sendHTTP({
      method: 'Service.ExecLivestream',
      params: [{
        URL: url
      }]
    })
  }

  public progressLivestream() {
    return this.sendHTTP<LiveStreamProgress>({
      method: 'Service.ProgressLivestream',
      params: []
    })
  }

  public killLivestream(url: string) {
    return this.sendHTTP({
      method: 'Service.KillLivestream',
      params: [url]
    })
  }

  public killAllLivestream() {
    return this.sendHTTP({
      method: 'Service.KillAllLivestream',
      params: []
    })
  }

  public updateExecutable() {
    return this.sendHTTP({
      method: 'Service.UpdateExecutable',
      params: []
    })
  }

  public clearCompleted() {
    return this.sendHTTP({
      method: 'Service.ClearCompleted',
      params: []
    })
  }
}