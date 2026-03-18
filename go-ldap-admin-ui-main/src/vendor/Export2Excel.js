/* eslint-disable */
import { saveAs } from 'file-saver'
import ExcelJS from 'exceljs'

function generateArray(table) {
  var out = [];
  var rows = table.querySelectorAll('tr');
  var ranges = [];
  for (var R = 0; R < rows.length; ++R) {
    var outRow = [];
    var row = rows[R];
    var columns = row.querySelectorAll('td');
    for (var C = 0; C < columns.length; ++C) {
      var cell = columns[C];
      var colspan = cell.getAttribute('colspan');
      var rowspan = cell.getAttribute('rowspan');
      var cellValue = cell.innerText;
      if (cellValue !== "" && cellValue == +cellValue) cellValue = +cellValue;

      //Skip ranges
      ranges.forEach(function(range) {
        if (R >= range.s.r && R <= range.e.r && outRow.length >= range.s.c && outRow.length <= range.e.c) {
          for (var i = 0; i <= range.e.c - range.s.c; ++i) outRow.push(null);
        }
      });

      //Handle Row Span
      if (rowspan || colspan) {
        rowspan = rowspan || 1;
        colspan = colspan || 1;
        ranges.push({
          s: {
            r: R,
            c: outRow.length
          },
          e: {
            r: R + rowspan - 1,
            c: outRow.length + colspan - 1
          }
        });
      };

      //Handle Value
      outRow.push(cellValue !== "" ? cellValue : null);

      //Handle Colspan
      if (colspan)
        for (var k = 0; k < colspan - 1; ++k) outRow.push(null);
    }
    out.push(outRow);
  }
  return [out, ranges];
};

export async function export_table_to_excel(id) {
  var theTable = document.getElementById(id);
  var oo = generateArray(theTable);
  var ranges = oo[1];
  var data = oo[0];

  const workbook = new ExcelJS.Workbook();
  const worksheet = workbook.addWorksheet('SheetJS');

  data.forEach(row => worksheet.addRow(row));

  ranges.forEach(range => {
    worksheet.mergeCells(
      range.s.r + 1, range.s.c + 1,
      range.e.r + 1, range.e.c + 1
    );
  });

  const buffer = await workbook.xlsx.writeBuffer();
  saveAs(
    new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }),
    'test.xlsx'
  );
}

export async function export_json_to_excel({
  multiHeader = [],
  header,
  data,
  filename,
  merges = [],
  autoWidth = true,
  bookType = 'xlsx'
} = {}) {
  filename = filename || 'excel-list'
  data = [...data]

  const workbook = new ExcelJS.Workbook();
  const worksheet = workbook.addWorksheet('SheetJS');

  for (const mh of multiHeader) {
    worksheet.addRow(mh);
  }
  worksheet.addRow(header);
  data.forEach(row => worksheet.addRow(row));

  if (merges.length > 0) {
    merges.forEach(item => worksheet.mergeCells(item));
  }

  if (autoWidth) {
    const allRows = [...multiHeader, header, ...data];
    worksheet.columns.forEach((column, colIndex) => {
      let maxLength = 10;
      allRows.forEach(row => {
        const val = row[colIndex];
        if (val == null) return;
        /*判断是否为中文*/
        const len = val.toString().charCodeAt(0) > 255
          ? val.toString().length * 2
          : val.toString().length;
        if (len > maxLength) maxLength = len;
      });
      column.width = maxLength;
    });
  }

  const buffer = await workbook.xlsx.writeBuffer();
  const mimeType = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';
  saveAs(new Blob([buffer], { type: mimeType }), `${filename}.${bookType}`);
}
