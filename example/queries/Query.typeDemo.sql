-- @one
SELECT {{param.id}}::uuid AS id,
       {{param.document}}::jsonb AS document,
       {{param.raw}}::json AS raw,
       {{param.ids}}::uuid[] AS ids,
       {{param.labels}}::text[] AS labels,
       {{param.counts}}::int[] AS counts,
       {{param.flags}}::boolean[] AS flags,
       {{param.scores}}::double precision[] AS scores,
       {{param.documents}}::jsonb[] AS documents,
       NULL::uuid AS "optionalId",
       NULL::jsonb AS "optionalDocument",
       NULL::text[] AS "optionalLabels",
       ARRAY[]::int[] AS "emptyCounts"
