<?php

namespace App\Controller\Admin;

use App\Entity\Address;
use App\Entity\User;
use EasyCorp\Bundle\EasyAdminBundle\Config\Action;
use EasyCorp\Bundle\EasyAdminBundle\Config\Actions;
use EasyCorp\Bundle\EasyAdminBundle\Config\Crud;
use EasyCorp\Bundle\EasyAdminBundle\Controller\AbstractCrudController;
use EasyCorp\Bundle\EasyAdminBundle\Router\CrudUrlGenerator;
use EasyCorp\Bundle\EasyAdminBundle\Context\AdminContext;

class UserCrudController extends AbstractCrudController
{
    public static function getEntityFqcn(): string
    {
        return User::class;
    }

    public function configureActions(Actions $actions): Actions
    {
        $sendInvoice = Action::new('sendInvoice', 'Address', 'fa fa-address-book')
            ->linkToCrudAction('renderInvoice')
            ->displayIf(fn ($entity) => $entity->getAddresses()->count() > 0);

        return $actions
            ->add(Crud::PAGE_INDEX, $sendInvoice);
    }

    public function renderInvoice(AdminContext $context)
    {
        $user = $context->getEntity()->getInstance();
        $id = $context->getRequest()->query->get('entityId');

        $adminUrlGenerator = $this->get(CrudUrlGenerator::class);

        $url = $adminUrlGenerator->build()
            ->setController(AddressCrudController::class)
            ->setAction(Action::INDEX)
            ->generateUrl();

        return $this->redirect($url);
    }

    /*
    public function configureFields(string $pageName): iterable
    {
        return [
            IdField::new('id'),
            TextField::new('title'),
            TextEditorField::new('description'),
        ];
    }
    */
}
