export interface Tag {
  id: string;
  name: string;
  color: string;
}

export interface TagPayload {
  name: string;
  color?: string;
}
