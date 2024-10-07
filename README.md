# Bootstrap Admin CRUD Auth Boilerplate with Golang Fiber

A full-stack boilerplate for both frontend and backend using Golang Fiber, designed for admin systems with basic features including authentication, user management, and project settings. It comes equipped with features like daily log management, file upload, file deletion, and table display using datatables with export options to PDF, CSV, Excel, and more.

This project is powered by Core UI (https://coreui.io/) for a fast and efficient frontend development experience.

## Getting Started

### 1. Configure Environment Variables
Create a `.env` file in the root directory of the project and fill it with your configuration settings with basic values from `.example.env`.

### 2. Install Dependencies
Run the following command to ensure all necessary modules are installed:

```bash
go mod tidy
```

### 3. Start the Development Server
To start the development server, run:

```bash
go run main.go
```

This will start the server and automatically load changes when you rerun the command after making changes.

Alternatively, if you prefer using air for live reloading during development, simply run:

```bash
air
```

Make sure to configure air according to your project's needs by adjusting the settings in the .air.toml file.

### 4. Start the Production Server
To start the server in production mode, you can build the binary and run it:

#### On Windows:
```bash
go build -o lorem.exe
lorem.exe
```

#### On Linux/macOS:
```bash
go build -o lorem
./lorem
```

# Bootstrap Admin CRUD Auth Boilerplate with Golang Fiber

A full-stack boilerplate for both frontend and backend using Golang Fiber, designed for admin systems with basic features including authentication, user management, and project settings. It comes equipped with features like daily log management, file upload, file deletion, and table display using datatables with export options to PDF, CSV, Excel, and more.

This project is powered by **Core UI** (https://coreui.io/) for a fast and efficient frontend development experience.

## Getting Started

### 1. Configure Environment Variables
Create a `.env` file in the root directory of the project and fill it with your configuration settings using the values from `.example.env`.

### 2. Install Dependencies
Run the following command to ensure all necessary modules are installed:

```bash
go mod tidy
```

### 3. Create Necessary Directories
Before running the server, ensure that the following directory exists for file uploads:

```bash
web/uploads/logs/
```

If the directory does not exist, create it manually:

#### On Windows:
```bash
mkdir web\uploads\logs
```

#### On Linux/macOS:
```bash
mkdir -p web/uploads/logs
```

This directory will be used to store user-uploaded files in the project’s daily logs.

### 4. Start the Development Server
To start the development server, run:

```bash
go run main.go
```

This will start the server and automatically reload changes when you rerun the command after making updates.

Alternatively, for live reloading during development, you can use **air** by running:

```bash
air
```

Make sure to configure air according to your project's needs by adjusting the settings in the `.air.toml` file.

### 5. Start the Production Server
To start the server in production mode, you can build the binary and run it:

#### On Windows:
```bash
go build -o lorem.exe
lorem.exe
```

#### On Linux/macOS:
```bash
go build -o lorem
./lorem
```

## Features

- **Authentication & Authorization:** Role-based access control, ensuring restricted access to various parts of the application.
- **User Management:** Administrators can add, update, or delete users, as well as assign different roles with delete implementation for this project using soft delete.
- **Project Management System:** Manage projects along with daily logs, including the ability to upload and delete files.
- **Data Table Management:** Display project and user data in a clean table format with export functionality to **PDF, CSV, Excel**, etc.
- **File Uploads:** Users can upload files related to daily logs, which are stored and managed efficiently.
- **Simple Dashboard**: 
  The system features a user-friendly **dashboard** with three core sections designed for simplicity and ease of use:
  
  1. **Main Dashboard**: This serves as the overview page, providing a quick glance at the overall data, such as recent activity within projects and daily logs.
  
  2. **Project Dashboard**: This page is dedicated to managing specific projects. Users can view all active projects, their statuses, and any relevant details. It also allows for easy navigation between different projects and their associated data.
  
  3. **Project Detail Dashboard**: A deeper dive into individual projects, this page includes detailed information such as **daily logs** tracking income and expenses. The logs are categorized and organized for clarity, with easy access to manage, update, and review financials or other key details related to the project with some simple chart.

## Folder Structure

The project is organized into a clear and easy-to-understand folder structure, facilitating modular development:

- `internal/`:  
  Contains core backend logic, including business logic, services, and utilities.
  
- `pkg/`:  
  Holds shared utilities or packages that can be reused across different parts of the project. Typically, this includes helpers and validation logic.
  
- `web/`:  
  Holds the frontend files, along with the location for user-uploaded files.
  - `web/uploads/logs/`:  
    The default folder where files uploaded by users in daily logs are stored.
  - `web/components/`:  
    Contains reusable frontend components, such as buttons, modals, and tables, used throughout the website.
  - `web/pages/`:  
    Main folder for the web pages that form the user interface of the application, including the login page, dashboard, and admin pages.

## Additional Information

- **Core UI** is used as the base template to speed up frontend development and provide a professional admin dashboard look.
- Default file structure includes a dedicated folder for logs (`web/uploads/logs`), which acts as the default storage for files uploaded by users in the daily logs.
- **File Deletion**: Users can manage uploaded files and delete them if needed from the logs section.
