export interface Profile {
  description: string;
  videoHeight: string;
  videoBitrate: string;
  videoBufsize: string;
  videoPreset: string;
  audioBitrate: string;
  audioProfile: string;
}

export interface TuneData {
  channel: string;
  profile: Profile | null;
}
