import * as z from "zod";

export const SubtypeSchema = z.enum([
  "Click",
  "DOMMutation",
  "FormInput",
  "FullSnapshot",
  "MouseMove",
  "Scroll",
  "ViewportChange",
]);
export type Subtype = z.infer<typeof SubtypeSchema>;

export const EventSchemasSchema = z.object({
  environment: z.string(),
  payload: z.record(z.string(), z.any()),
  releaseId: z.string(),
  sessionId: z.string(),
  subtype: SubtypeSchema,
});
export type EventSchemas = z.infer<typeof EventSchemasSchema>;
