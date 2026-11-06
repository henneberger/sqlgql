-- @one
SELECT {{param.big}}::bigint AS big,
       {{param.real}}::real AS real,
       {{param.double}}::double precision AS double,
       {{param.bigs}}::bigint[] AS bigs,
       {{param.reals}}::real[] AS reals,
       {{param.doubles}}::double precision[] AS doubles,
       NULL::bigint AS "optionalBig",
       NULL::real AS "optionalReal",
       NULL::double precision AS "optionalDouble",
       CASE {{param.special}}::text
         WHEN 'nan' THEN 'NaN'::double precision
         WHEN 'infinity' THEN 'Infinity'::double precision
         WHEN '-infinity' THEN '-Infinity'::double precision
         ELSE 0::double precision
       END AS special
