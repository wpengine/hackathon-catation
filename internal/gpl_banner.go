// Copyright (C) 2020  WPEngine
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package internal

import "fmt"

func PrintGPLBanner(program, releaseYear string) {
	fmt.Printf(
		`%s  Copyright (C) %s  WP Engine
This program comes with ABSOLUTELY NO WARRANTY; for details, see https://www.gnu.org/licenses/gpl-3.0.txt.
This is free software, and you are welcome to redistribute it
under certain conditions; for details, see https://www.gnu.org/licenses/gpl-3.0.txt.

`, program, releaseYear)
}
