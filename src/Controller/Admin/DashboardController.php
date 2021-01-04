<?php

namespace App\Controller\Admin;

use EasyCorp\Bundle\EasyAdminBundle\Config\Dashboard;
use EasyCorp\Bundle\EasyAdminBundle\Config\MenuItem;
use EasyCorp\Bundle\EasyAdminBundle\Controller\AbstractDashboardController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;
use App\Entity\Address;
use App\Entity\Category;
use App\Entity\Order;
use App\Entity\Product;
use App\Entity\User;

class DashboardController extends AbstractDashboardController
{
    /**
     * @Route("/admin", name="admin")
     */
    public function index(): Response
    {
        return parent::index();
    }

    public function configureDashboard(): Dashboard
    {
        return Dashboard::new()
            ->setTitle('Panzucha');
    }

    public function configureMenuItems(): iterable
    {
        yield MenuItem::linktoDashboard('Dashboard', 'fa fa-home');        
        yield MenuItem::linkToCrud('Category', 'fas fa-chevron-circle-right', Category::class);
        yield MenuItem::linkToCrud('Product', 'fas fa-chevron-circle-right', Product::class);
        yield MenuItem::linkToCrud('User', 'fas fa-user', User::class);
        yield MenuItem::linkToCrud('Order', 'fas fa-user', Order::class);
        //return [
            // ...
          yield  MenuItem::linkToLogout('Logout', 'fas fa-user-slash');
        //];
    }
}
