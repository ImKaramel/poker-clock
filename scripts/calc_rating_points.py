#!/usr/bin/env python3
"""CLI для расчёта очков рейтинга по числу игроков и нокаутам.

Базовая шкала взята из эталонного турнира на 55 игроков.
Для любого размера поля очки за место масштабируются по формуле:

    round(base_points_for_place * players_count / 55)

Если игроков больше 55, для мест после 55 используется минимальная
базовая ставка из эталонной таблицы.

Примеры:
    python scripts/calc_rating_points.py --players 10 20 30
    python scripts/calc_rating_points.py --results results.json
    python scripts/calc_rating_points.py --results results.json --ko-bonus 100
"""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

BASELINE_POINTS = [
    3100.00,
    2738.43,
    2233.03,
    2108.00,
    1983.11,
    1612.00,
    1378.38,
    1351.35,
    1240.00,
    1219.51,
    1161.17,
    1116.00,
    1013.51,
    992.00,
    927.89,
    878.05,
    868.00,
    829.27,
    829.27,
    784.62,
    744.00,
    702.70,
    664.54,
    634.15,
    634.15,
    620.00,
    527.03,
    496.00,
    487.80,
    487.80,
    486.49,
    461.54,
    439.02,
    439.02,
    405.41,
    405.41,
    390.24,
    390.24,
    372.00,
    364.86,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
    124.00,
]
BASELINE_FIELD_SIZE = 55
FALLBACK_PLACE_POINTS = BASELINE_POINTS[-1]


def scaled_place_points(players_count: int) -> list[int]:
    if players_count <= 0:
        raise ValueError("players_count must be positive")

    table: list[int] = []
    for position in range(1, players_count + 1):
        base_value = (
            BASELINE_POINTS[position - 1]
            if position <= len(BASELINE_POINTS)
            else FALLBACK_PLACE_POINTS
        )
        table.append(round(base_value * players_count / BASELINE_FIELD_SIZE))
    return table


def render_place_table(players_count: int) -> str:
    lines = [f"=== {players_count} players ==="]
    points = scaled_place_points(players_count)
    for position, score in enumerate(points, start=1):
        lines.append(f"{position:>2}. {score}")
    return "\n".join(lines)


def load_results(path: Path) -> list[dict[str, Any]]:
    payload = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(payload, list):
        raise ValueError("results JSON must be a list")
    return payload


def compute_results(
    results: list[dict[str, Any]],
    players_count: int,
    ko_bonus: int,
) -> list[dict[str, Any]]:
    place_points = scaled_place_points(players_count)
    output: list[dict[str, Any]] = []

    for row in results:
        name = str(row.get("name", "")).strip()
        position = int(row["position"])
        ko_count = int(row.get("ko", 0))
        manual_bonus = int(row.get("bonus", 0))
        if position < 1 or position > players_count:
            raise ValueError(f"position {position} is outside 1..{players_count}")

        base_score = place_points[position - 1]
        ko_score = ko_count * ko_bonus
        total = base_score + ko_score + manual_bonus
        output.append(
            {
                "name": name or f"Player {position}",
                "position": position,
                "base_points": base_score,
                "ko": ko_count,
                "ko_points": ko_score,
                "bonus": manual_bonus,
                "total_points": total,
            }
        )

    output.sort(key=lambda item: item["position"])
    return output


def render_results(results: list[dict[str, Any]], players_count: int, ko_bonus: int) -> str:
    lines = [
        f"=== Rating results for {players_count} players ===",
        f"KO bonus: {ko_bonus}",
        "",
        "place | name | base | ko | bonus | total",
    ]
    for row in results:
        lines.append(
            f"{row['position']:>5} | {row['name']} | {row['base_points']:>4} | "
            f"{row['ko_points']:>3} ({row['ko']}) | {row['bonus']:>5} | {row['total_points']:>5}"
        )
    return "\n".join(lines)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Расчёт очков рейтинга Midnight Club")
    parser.add_argument(
        "--players",
        type=int,
        nargs="+",
        help="Количество игроков. Можно передать несколько значений: --players 10 20 30",
    )
    parser.add_argument(
        "--results",
        type=Path,
        help="JSON-файл с результатами: [{\"name\":\"Player\",\"position\":1,\"ko\":2,\"bonus\":100}]",
    )
    parser.add_argument(
        "--ko-bonus",
        type=int,
        default=100,
        help="Сколько очков даёт один knockout. По умолчанию 100.",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Печатать итог в JSON для --results.",
    )
    return parser


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()

    if not args.players and not args.results:
        parser.error("нужно передать --players и/или --results")

    if args.results:
        try:
            raw_results = load_results(args.results)
            players_count = args.players[0] if args.players else len(raw_results)
            output = compute_results(raw_results, players_count, args.ko_bonus)
        except (FileNotFoundError, KeyError, ValueError) as exc:
            parser.error(str(exc))

        if args.json:
            print(json.dumps(output, ensure_ascii=False, indent=2))
            return
        print(render_results(output, players_count, args.ko_bonus))
        return

    for index, players_count in enumerate(args.players, start=1):
        if index > 1:
            print()
        print(render_place_table(players_count))


if __name__ == "__main__":
    main()
