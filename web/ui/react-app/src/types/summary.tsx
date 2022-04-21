export interface MonitorSummaryType {
  service: ServiceSummaryListType;
  order: string[];
}
export interface ServiceSummaryListType {
  [id: string]: ServiceSummaryType;
}

export interface ServiceSummaryType {
  id: string;
  loading: boolean;
  type?: string;
  url?: string;
  icon?: string;
  gotify?: boolean;
  slack?: boolean;
  webhook?: number;
  status?: StatusSummaryType;
}

export interface WebHookModal {
  type: ModalType;
  service: ServiceSummaryType;
}

export type ModalType = "RESEND" | "RETRY" | "SEND" | "SKIP" | "";

export interface WebHookModalData {
  service_id: string;
  sent: string[];
  webhooks: WebHookSummaryListType;
}

export interface StatusSummaryType {
  approved_version?: string;
  current_version?: string;
  current_version_timestamp?: string;
  latest_version?: string;
  latest_version_timestamp?: string;
  last_queried?: string;
  // fails?: StatusFailsSummaryType;
}

export interface StatusFailsSummaryType {
  gotify?: boolean;
  slack?: boolean;
  webhook?: boolean;
}

export interface WebHookSummaryType {
  // undefined = unsent/sending
  failed?: boolean;
}

export interface WebHookSummaryListType {
  [id: string]: WebHookSummaryType;
}