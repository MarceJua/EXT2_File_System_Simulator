# EXT2 File System Simulator

## Overview
This project implements a simulation of the **EXT2 file system** as part of the *Laboratorio Manejo e Implementación de Archivos* course (Section B, 2025). Developed by **Marcelo Andre Juarez Alfaro (202010367)**, it features a client-server architecture with a **Go backend** using the Fiber framework and a **Next.js frontend**, communicating via a RESTful API. The system supports core EXT2 file system operations, such as disk creation, partitioning, formatting, and user/group management, with data stored in `.mia` binary files.

For a comprehensive technical breakdown, refer to the [Technical Manual](Manual%20Técnico%20-%20Proyecto%201_%20Sistema%20de%20Archivos%20EXT2.pdf) in this repository.

## Features
- **Disk Management**: Create (`MKDISK`), delete (`RMDISK`), and partition (`FDISK`) virtual disks stored as `.mia` files.
- **Partition Management**: Mount (`MOUNT`) and list (`MOUNTED`) partitions, with support for primary, extended, and logical partitions.
- **File System Operations**: Format partitions with EXT2 (`MKFS`), create directories (`MKDIR`), and manage files (`MKFILE`, `CAT`).
- **User and Group Management**: Create (`MKUSR`, `MKGRP`), delete (`RMUSR`, `RMGRP`), and modify (`CHGRP`) users and groups, with session handling (`LOGIN`, `LOGOUT`).
- **Reporting**: Generate Graphviz-based reports (`REP`) for structures like MBR, Superblock, and more.
- **Interactive Interface**: A Next.js-based web frontend allows users to input commands or upload scripts, with results displayed in real-time.

## Technologies
- **Frontend**: Next.js with React for a dynamic, client-side rendered interface.
- **Backend**: Go with Fiber framework for a high-performance RESTful API.
- **File System**: Simulated EXT2 structures (MBR, Superblock, Inodes, Bitmaps, etc.) stored in `.mia` files.
- **Communication**: HTTP-based RESTful API for frontend-backend interaction.
- **Utilities**: Custom Go packages (`utils`, `stores`) for disk operations, serialization, and session management.

## Setup Instructions
1. **Prerequisites**:
   - Node.js (v16 or higher) for the frontend.
   - Go (v1.18 or higher) for the backend.
   - Git for cloning the repository.

2. **Clone the Repository**:
   ```bash
   git clone https://github.com/MarceJua/EXT2_FileSystem_Simulator.git
   cd EXT2_FileSystem_Simulator
   ```

3. **Backend Setup**:
   - Navigate to the backend directory:
     ```bash
     cd backend
     ```
   - Install dependencies:
     ```bash
     go mod tidy
     ```
   - Run the server:
     ```bash
     go run main.go
     ```
     The backend will start on `http://localhost:3001`.

4. **Frontend Setup**:
   - Navigate to the frontend directory:
     ```bash
     cd frontend
     ```
   - Install dependencies:
     ```bash
     npm install
     ```
   - Run the development server:
     ```bash
     npm run dev
     ```
     The frontend will be available at `http://localhost:3000`.

5. **Usage**:
   - Access the web interface at `http://localhost:3000`.
   - Use the input terminal or file upload feature to execute commands (e.g., `mkdisk -size=10 -unit=M -path=/home/disco.mia`).
   - View results in the output terminal.

## Documentation
For detailed information on the architecture, data structures (MBR, Superblock, Inodes, etc.), and command implementations, consult the [Technical Manual](Manual%20Técnico%20-%20Proyecto%201_%20Sistema%20de%20Archivos%20EXT2.pdf) included in this repository.

## Contributing
This project is part of an academic assignment and is not currently open to contributions. For feedback or inquiries, contact the author at [mjuarez2017ig@gmail.com].
