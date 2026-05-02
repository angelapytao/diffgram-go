#!/usr/bin/env python3
"""
pg_to_mysql.py — translate Diffgram PG 15 schema dump to MySQL 8 DDL.

Usage:
    python scripts/pg-to-mysql/pg_to_mysql.py \
        /path/to/diffgram_complete_schema.sql \
        migrations/00001_initial_schema.sql
"""

import re
import sys
from pathlib import Path

TYPE_MAP: dict[str, str] = {
    "bigint": "BIGINT",
    "integer": "INT",
    "smallint": "SMALLINT",
    "double precision": "DOUBLE",
    "real": "FLOAT",
    "boolean": "TINYINT(1)",
    "text": "TEXT",
    "jsonb": "JSON",
    "json": "JSON",
    "bytea": "BLOB",
    "date": "DATE",
    "time without time zone": "TIME",
    "timestamp without time zone": "DATETIME(6)",
    "timestamp with time zone": "DATETIME(6)",
}

MULTI_WORD = sorted(TYPE_MAP.keys(), key=len, reverse=True)


def convert_type(pg: str) -> str:
    pg = pg.strip()
    if "[]" in pg:
        return "JSON"
    m = re.match(r"character varying\((\d+)\)", pg)
    if m:
        return f"VARCHAR({m.group(1)})"
    if pg in ("character varying", "varchar"):
        return "VARCHAR(255)"
    m = re.match(r"character\((\d+)\)", pg)
    if m:
        return f"CHAR({m.group(1)})"
    m = re.match(r"numeric\((\d+),\s*(\d+)\)", pg)
    if m:
        return f"DECIMAL({m.group(1)},{m.group(2)})"
    return TYPE_MAP.get(pg, pg.upper())


def strip_cast(val: str) -> str:
    return re.sub(r"::[a-z ]+(\[\])?", "", val).strip()


def parse_column_line(raw: str, pk_col: str = "id") -> str | None:
    raw = raw.strip().rstrip(",")
    if not raw or raw.startswith("--"):
        return None
    if re.match(r"CONSTRAINT\b", raw, re.IGNORECASE):
        return None

    m = re.match(r"^(\w+)\s+(.*)", raw, re.DOTALL)
    if not m:
        return None
    col_name = m.group(1)
    rest = m.group(2).strip()

    mysql_type: str | None = None
    remaining = rest

    for pg_type in MULTI_WORD:
        if rest.lower().startswith(pg_type):
            after = rest[len(pg_type):]
            if pg_type == "character varying":
                mn = re.match(r"\((\d+)\)(.*)", after, re.DOTALL)
                if mn:
                    mysql_type = f"VARCHAR({mn.group(1)})"
                    after = mn.group(2)
                else:
                    mysql_type = "VARCHAR(255)"
            else:
                mysql_type = TYPE_MAP[pg_type]
            after = after.lstrip()
            if after.startswith("[]"):
                mysql_type = "JSON"
                after = after[2:]
            remaining = after.lstrip()
            break

    if mysql_type is None:
        m2 = re.match(r"(\w+)(\(\d+(?:,\s*\d+)?\))?(\[\])?(.*)", rest, re.DOTALL)
        if not m2:
            return None
        raw_type = m2.group(1).lower()
        size_part = m2.group(2) or ""
        is_array = bool(m2.group(3))
        remaining = (m2.group(4) or "").lstrip()

        if is_array:
            mysql_type = "JSON"
        elif raw_type == "character" and size_part:
            mysql_type = f"CHAR({size_part[1:-1]})"
        elif raw_type in ("varchar", "character"):
            mysql_type = "VARCHAR(255)"
        else:
            mysql_type = TYPE_MAP.get(raw_type, raw_type.upper())
            if size_part and mysql_type not in ("TEXT", "JSON", "BLOB", "DATETIME(6)"):
                mysql_type = f"{mysql_type}{size_part}"

    not_null = False
    default_val: str | None = None

    remaining = remaining.strip()
    while remaining:
        up = remaining.upper()
        if up.startswith("NOT NULL"):
            not_null = True
            remaining = remaining[8:].strip()
        elif up.startswith("DEFAULT"):
            remaining = remaining[7:].strip()
            m3 = re.match(r"^(.+?)(\s+NOT NULL)?\s*$", remaining, re.DOTALL)
            if m3:
                raw_default = m3.group(1).strip()
                default_val = strip_cast(raw_default)
                if m3.group(2):
                    not_null = True
            remaining = ""
        else:
            break

    result = f"  `{col_name}` {mysql_type}"
    if col_name == pk_col and not_null:
        result += " NOT NULL AUTO_INCREMENT"
    else:
        if not_null:
            result += " NOT NULL"
        if default_val is not None:
            result += f" DEFAULT {default_val}"

    return result


def parse_tables(src: str) -> dict[str, list[str]]:
    tables: dict[str, list[str]] = {}
    for m in re.finditer(
        r"CREATE TABLE public\.(\w+)\s*\(\n(.*?)\n\);",
        src, re.DOTALL,
    ):
        body = m.group(2)
        tables[m.group(1)] = [ln for ln in body.splitlines()]
    return tables


def parse_pks(src: str) -> dict[str, list[str]]:
    pk_map: dict[str, list[str]] = {}
    for m in re.finditer(
        r"ALTER TABLE ONLY public\.(\w+)\s+ADD CONSTRAINT \w+ PRIMARY KEY \(([^)]+)\);",
        src,
    ):
        pk_map[m.group(1)] = [c.strip() for c in m.group(2).split(",")]
    return pk_map


def parse_fks(src: str) -> dict[str, list[dict]]:
    fk_map: dict[str, list[dict]] = {}
    for m in re.finditer(
        r"ALTER TABLE ONLY public\.(\w+)\s+"
        r"ADD CONSTRAINT (\w+) FOREIGN KEY \(([^)]+)\)\s*"
        r"REFERENCES public\.(\w+)\s*\(([^)]+)\)[^;]*;",
        src,
    ):
        fk_map.setdefault(m.group(1), []).append({
            "name": m.group(2),
            "cols": m.group(3).strip(),
            "ref_table": m.group(4),
            "ref_cols": m.group(5).strip(),
        })
    return fk_map


def parse_indexes(src: str) -> dict[str, list[str]]:
    idx_map: dict[str, list[str]] = {}
    for m in re.finditer(
        r"CREATE (UNIQUE )?INDEX (\w+) ON public\.(\w+) USING (\w+) \(([^)]+)\)"
        r"(?:\s+WHERE [^;]+)?;",
        src,
    ):
        unique = m.group(1) or ""
        idx_name = m.group(2)
        table = m.group(3)
        method = m.group(4).lower()
        cols = m.group(5)
        if method == "gin":
            continue
        col_list = ", ".join(f"`{c.strip()}`" for c in cols.split(","))
        stmt = f"CREATE {unique}INDEX `{idx_name}` ON `{table}` ({col_list});"
        idx_map.setdefault(table, []).append(stmt)
    return idx_map


def build_create_table(
    table: str,
    raw_lines: list[str],
    pk_cols: list[str],
    fks: list[dict],
) -> list[str]:
    pk_col = pk_cols[0] if len(pk_cols) == 1 else "id"

    column_defs: list[str] = []
    for raw in raw_lines:
        col = parse_column_line(raw, pk_col=pk_col)
        if col is not None:
            column_defs.append(col)

    if not column_defs:
        column_defs = ["  `id` INT NOT NULL AUTO_INCREMENT"]

    parts = list(column_defs)

    if pk_cols:
        pk_str = ", ".join(f"`{c}`" for c in pk_cols)
        parts.append(f"  PRIMARY KEY ({pk_str})")

    for fk in fks:
        col_str = ", ".join(f"`{c.strip()}`" for c in fk["cols"].split(","))
        ref_col_str = ", ".join(f"`{c.strip()}`" for c in fk["ref_cols"].split(","))
        parts.append(
            f"  CONSTRAINT `{fk['name']}` FOREIGN KEY ({col_str})"
            f" REFERENCES `{fk['ref_table']}` ({ref_col_str})"
        )

    lines = [f"CREATE TABLE IF NOT EXISTS `{table}` ("]
    for i, p in enumerate(parts):
        comma = "," if i < len(parts) - 1 else ""
        lines.append(p + comma)
    lines.append(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;")
    return lines


def main() -> None:
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <pg_dump.sql> <output.sql>", file=sys.stderr)
        sys.exit(1)

    src = Path(sys.argv[1]).read_text()
    out_path = Path(sys.argv[2])

    tables = parse_tables(src)
    pk_map = parse_pks(src)
    fk_map = parse_fks(src)
    idx_map = parse_indexes(src)

    out: list[str] = [
        "-- +goose Up",
        "-- Generated by scripts/pg-to-mysql/pg_to_mysql.py — DO NOT EDIT manually.",
        "SET FOREIGN_KEY_CHECKS=0;",
        "",
    ]

    drop_stmts: list[str] = []

    for table, raw_lines in tables.items():
        create_lines = build_create_table(
            table,
            raw_lines,
            pk_map.get(table, []),
            fk_map.get(table, []),
        )
        out += create_lines
        out.append("")
        for idx_stmt in idx_map.get(table, []):
            out.append(idx_stmt)
        if idx_map.get(table):
            out.append("")
        drop_stmts.append(f"DROP TABLE IF EXISTS `{table}`;")

    out += [
        "SET FOREIGN_KEY_CHECKS=1;",
        "",
        "-- +goose Down",
        "SET FOREIGN_KEY_CHECKS=0;",
        "",
    ]
    out += list(reversed(drop_stmts))
    out += ["", "SET FOREIGN_KEY_CHECKS=1;"]

    out_path.write_text("\n".join(out) + "\n")
    print(f"✓ translated {len(tables)} tables → {out_path}")


if __name__ == "__main__":
    main()
