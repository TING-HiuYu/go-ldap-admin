<template>
  <div>
    <input ref="excel-upload-input" class="excel-upload-input" type="file" accept=".xlsx, .xls" @change="handleClick">
    <div class="drop" @drop="handleDrop" @dragover="handleDragover" @dragenter="handleDragover">
      Drop excel file here or
      <el-button :loading="loading" style="margin-left:16px;" size="mini" type="primary" @click="handleUpload">
        Browse
      </el-button>
    </div>
  </div>
</template>

<script>
import ExcelJS from 'exceljs'

export default {
  props: {
    beforeUpload: Function, // eslint-disable-line
    onSuccess: Function// eslint-disable-line
  },
  data() {
    return {
      loading: false,
      excelData: {
        header: null,
        results: null
      }
    }
  },
  methods: {
    generateData({ header, results }) {
      this.excelData.header = header
      this.excelData.results = results
      this.onSuccess && this.onSuccess(this.excelData)
    },
    handleDrop(e) {
      e.stopPropagation()
      e.preventDefault()
      if (this.loading) return
      const files = e.dataTransfer.files
      if (files.length !== 1) {
        this.$message.error('Only support uploading one file!')
        return
      }
      const rawFile = files[0] // only use files[0]

      if (!this.isExcel(rawFile)) {
        this.$message.error('Only supports upload .xlsx, .xls, .csv suffix files')
        return false
      }
      this.upload(rawFile)
      e.stopPropagation()
      e.preventDefault()
    },
    handleDragover(e) {
      e.stopPropagation()
      e.preventDefault()
      e.dataTransfer.dropEffect = 'copy'
    },
    handleUpload() {
      this.$refs['excel-upload-input'].click()
    },
    handleClick(e) {
      const files = e.target.files
      const rawFile = files[0] // only use files[0]
      if (!rawFile) return
      this.upload(rawFile)
    },
    upload(rawFile) {
      this.$refs['excel-upload-input'].value = null // fix can't select the same excel

      if (!this.beforeUpload) {
        this.readerData(rawFile)
        return
      }
      const before = this.beforeUpload(rawFile)
      if (before) {
        this.readerData(rawFile)
      }
    },
    readerData(rawFile) {
      this.loading = true
      return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onload = async e => {
          try {
            const workbook = new ExcelJS.Workbook()
            await workbook.xlsx.load(e.target.result)
            const worksheet = workbook.getWorksheet(1)
            const header = this.getHeaderRow(worksheet)
            const results = this.getSheetData(worksheet, header)
            this.generateData({ header, results })
            this.loading = false
            resolve()
          } catch (err) {
            this.loading = false
            reject(err)
          }
        }
        reader.readAsArrayBuffer(rawFile)
      })
    },
    getHeaderRow(worksheet) {
      const headers = []
      const firstRow = worksheet.getRow(1)
      firstRow.eachCell({ includeEmpty: false }, (cell, colNumber) => {
        const val = cell.value
        headers.push(val !== null && val !== undefined ? String(val) : 'UNKNOWN ' + (colNumber - 1))
      })
      return headers
    },
    getSheetData(worksheet, header) {
      const results = []
      worksheet.eachRow((row, rowNumber) => {
        if (rowNumber === 1) return // skip header row
        const rowData = {}
        header.forEach((key, index) => {
          const cell = row.getCell(index + 1)
          rowData[key] = cell.value
        })
        results.push(rowData)
      })
      return results
    },
    isExcel(file) {
      return /\.(xlsx|xls|csv)$/.test(file.name)
    }
  }
}
</script>


<style scoped>
.excel-upload-input{
  display: none;
  z-index: -9999;
}
.drop{
  border: 2px dashed #bbb;
  width: 600px;
  height: 160px;
  line-height: 160px;
  margin: 0 auto;
  font-size: 24px;
  border-radius: 5px;
  text-align: center;
  color: #bbb;
  position: relative;
}
</style>
