// generated by berty.tech/berty/tool/tyber/gen

export type TyberStatusType = string

export interface TyberDetail {
  name: string
  description: string
}

export interface TyberStep {
  parentTraceID: string
  details: TyberDetail[]
  status: TyberStatusType
  endTrace: boolean
  updateTraceName: string
  forceReopen: boolean
}

export interface ParserAppStep {
  name: string
  details: TyberDetail[]
  status: TyberStatusType
  started: string
  finished: string
  forceReopen: boolean
  updateTraceName: string
}

export interface ParserCreateStepEvent {
  parentID: string
  name: string
  details: TyberDetail[]
  status: TyberStatusType
  started: string
  finished: string
  forceReopen: boolean
  updateTraceName: string
}

export interface ParserSubTarget {
  TargetName: string
  TargetDetails: TyberDetail[]
  StepToAdd: TyberStep
}

export interface ParserCreateTraceEvent {
  id: string
  name: string
  steps: ParserAppStep[]
  Subs: ParserSubTarget[]
  status: TyberStatusType
  started: string
  finished: string
}

export interface ParserUpdateTraceEvent {
  id: string
  name: string
  status: TyberStatusType
  started: string
  finished: string
}

